<html>
<head>
  <script src="Chart.min.js"></script>
</head>
<body>
  <div style="width:165px; height: 55px;">
    <canvas id="vitalityChart" width="165px" height="55px"></canvas>
  </div>
  <script>
  var ctx = document.getElementById("vitalityChart").getContext("2d");
    var myLineChart = new Chart(ctx, {
      type: 'line',
      data: {
          labels:{{ .Labels }},
          datasets: [{
          fill: false,
          lineTension: 0.1,
          data:{{ .VitalitySlice }},
          borderWidth: 2,
          borderColor: '#06c'
         }]
       },
      options: {
        scales: {
          xAxes: [{
            display: false,
            stacked: true
          }],
          yAxes: [{
            display: false,
            stacked: true,
            ticks: {
              beginAtZero:true
            }
          }]
        },
        legend: {
          display: false,
        },
        elements: {
          point: {
            radius: 0,
            hoverRadius: 0,
            hoverBorderWidth: 0
          }
        },
        layout: {
          padding: {
            left: 5,
            right: 5,
            top: 5,
            bottom: 5
          }
        },
        tooltips: {
          enabled: false
        }
      }
    });
  </script>
</body>
</html>
