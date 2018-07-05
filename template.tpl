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
        labels:['0','1','2','3','4','5','6','7','8','9','10','11','12','13','14','15','16','17','18','19','20','21','22','23','24','25','26','27','28','29','','31','32','33','34','35','36','37','38','39','40','41','42','43','44','45','46','47','48','49','50','51','52','53','54','55','56','57','58','59'],
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
