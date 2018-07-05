package main

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	yaml "gopkg.in/yaml.v2"
)

// RangesData contains the data loaded from ranges.yml
type RangesData []Ranges

// Ranges are the ranges for a specific parameter (userCommunity, codeActivity, releaseHistory, longevity).
type Ranges struct {
	Name   string
	Ranges []Range
}

// Range is a range between will be assigned Points value.
type Range struct {
	Min    float64
	Max    float64
	Points float64
}

// CalculateRepoActivity return the repository activity index and the vitality slice calculated on the git clone.
// It follows the document https://lg-acquisizione-e-riuso-software-per-la-pa.readthedocs.io/
// In reference to section 2.5: fase-2-2-valutazione-soluzioni-riusabili-per-la-pa
func CalculateRepoActivity(folder string, days int) (float64, map[int]float64, error) {
	log.Debugf("CalculateRepoActivity")
	// Folder cannot be empty.
	if folder == "" {
		return 0, nil, errors.New("cannot calculate repository activity without folder name")
	}

	// MkdirAll will create all the folder path, if not exists.
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return 0, nil, err
	}

	// Repository activity score.
	var (
		userCommunity  float64
		codeActivity   float64
		releaseHistory float64
		longevity      float64

		repoActivity float64
	)

	// Open and load the git repo path.
	log.Debugf("Plain open: %s", folder)
	r, err := git.PlainOpen(folder)
	if err != nil {
		log.Error(err)
	}

	// Extract all the commits.
	log.Debug("Extract all commits")
	commits, err := extractAllCommits(r)
	if err != nil {
		log.Error(err)
	}

	// List commits before a number of days: commitsLastDays[from days before today][]commits
	commitsLastDays := map[int][]*object.Commit{}
	// Populate the slice of commits in every day.
	for i := 0; i < days; i++ {
		lastDays := time.Now().AddDate(0, 0, -i)
		// Append all the commits created before lastDays date.
		for _, c := range commits {
			if c.Author.When.Before(lastDays) {
				commitsLastDays[i] = append(commitsLastDays[i], c)
			}
		}
	}

	// List commits in a day: commitsPerDay[from days before today][]commits
	commitsPerDay := map[int][]*object.Commit{}
	// Populate the slice of commits in every day.
	for i := 0; i < days; i++ {
		lastDays := time.Now().AddDate(0, 0, -i)
		// Append all the commits created before lastDays date.
		for _, c := range commits {
			if c.Author.When.Day() == lastDays.Day() && c.Author.When.Month() == lastDays.Month() && c.Author.When.Year() == lastDays.Year() {
				commitsPerDay[i] = append(commitsPerDay[i], c)
			}
		}
	}

	// Extract all tags.
	log.Debug("Extract all Tags")
	tags, err := extractAllTagsCommit(r)
	if err != nil {
		log.Error(err)
	}
	tagsPerDays := map[int][]*object.Commit{}
	// Populate the slice of commits in every day.
	for i := 0; i < days; i++ {
		lastDays := time.Now().AddDate(0, 0, -i)
		// Append all the commits created before lastDays date.
		for _, t := range tags {
			if t != nil {
				if t.Author.When.Day() == lastDays.Day() && t.Author.When.Month() == lastDays.Month() && t.Author.When.Year() == lastDays.Year() {
					tagsPerDays[i] = append(tagsPerDays[i], t)
				}
			}
		}
	}

	// For every day (and before) calculate the Vitality index.
	vitalityIndex := map[int]float64{}

	// Longevity is the repository age.
	log.Debug("Calculate longevity")
	longevity, err = calculateLongevityIndex(r)
	if err != nil {
		log.Warn(err)
	}

	log.Debugf("Calculating userCommunity, codeActivity, releaseHistory and repoActivity for %d days before Today().", days)
	for i := 0; i < days; i++ {
		userCommunity = ranges("userCommunity", userCommunityLastDays(commitsLastDays[i]))

		codeActivity = ranges("codeActivity", activityLastDays(commitsPerDay[i]))
		releaseHistory = ranges("releaseHistory", releaseHistoryLastDays(tagsPerDays[i]))

		repoActivity = float64(userCommunity) + float64(codeActivity) + float64(releaseHistory) + ranges("longevity", float64(longevity))
		vitalityIndex[i] = repoActivity
	}

	return vitalityIndex[0], vitalityIndex, nil
}

// userCommunityLastDays returns the number of unique commits authors.
func userCommunityLastDays(commits []*object.Commit) float64 {
	// Prepare single author map.
	totalAuthors := map[string]int{}
	// Iterates over the commits and extract infos.
	for _, c := range commits {
		totalAuthors[c.Author.Email]++
	}
	return float64(len(totalAuthors))
}

// activityLastDays: # commits and # merges
func activityLastDays(commits []*object.Commit) float64 {
	numberCommits := float64(len(commits))
	numberMerges := 0
	for _, c := range commits {
		if c.NumParents() > 1 {
			numberMerges++
		}
	}

	return numberCommits + float64(numberMerges)
}

// releaseHistoryLastDays: number of releases
func releaseHistoryLastDays(tags []*object.Commit) float64 {
	return float64(len(tags))
}

// Extract all commits referred to released Tags.
func extractAllTagsCommit(r *git.Repository) ([]*object.Commit, error) {
	var allTags []*object.Commit

	tagrefs, err := r.Tags()
	if err != nil {
		return nil, err
	}
	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		if !t.Hash().IsZero() {
			tagObject, _ := r.CommitObject(t.Hash())
			allTags = append(allTags, tagObject)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return allTags, nil
}

func extractAllCommits(r *git.Repository) ([]*object.Commit, error) {
	var commits []*object.Commit

	ref, err := r.Head()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = cIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return commits, nil
}

// calculateLongevityIndex
func calculateLongevityIndex(r *git.Repository) (float64, error) {
	ref, err := r.Head()
	if err != nil {
		log.Error(err)
		return 0, err
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		log.Error(err)
		return 0, err
	}
	creationDate, err := extractOldestCommitDate(cIter)
	if err != nil {
		log.Error(err)
	}

	age := float64(time.Since(creationDate).Hours()) / 24

	// Git was invented in 2005. If some repo starts before, remove.
	then := time.Date(2005, time.January, 1, 1, 0, 0, 0, time.UTC)
	duration := time.Since(then).Hours()
	if age > duration/24 {
		return -1, errors.New("first commit is too old. Must be after the creation of git (2005)")
	}

	return age, err
}

// extractOldestCommitDate returns the oldest commit date.
func extractOldestCommitDate(cIter object.CommitIter) (time.Time, error) {
	// Iterates over the commits and extract infos.
	result := time.Now()
	err := cIter.ForEach(func(c *object.Commit) error {
		if c.Author.When.Before(result) {
			result = c.Author.When
		}
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	return result, nil
}

func ranges(name string, value float64) float64 {
	data, err := ioutil.ReadFile("ranges.yml")
	if err != nil {
		log.Error(err)
	}
	// Prepare the data structure for load.
	t := RangesData{}
	// Populate the yaml.
	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		log.Errorf("error: %v", err)
	}

	for _, v := range t {
		// Select the right ranges table.
		if v.Name == name {
			for _, r := range v.Ranges {
				if value >= r.Min && value < r.Max {
					return r.Points
				}
			}

		}
	}

	return 0
}
