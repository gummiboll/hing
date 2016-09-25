$(function() {
  var statistics = {
          url: "",
          numRequests: 0,
          statusCodes: {},
          errors: {},
          rts: [],
          Codes: function() {
                  tmp = {};
                  for (var key in statistics.statusCodes) {
                         tmp[key] = statistics.statusCodes[key];
                  }
                  return tmp;
          },
          NumRequests: function() {
                  return statistics.rts.length;
          },
          Calculated: function() {
                var tot = 0;
                var slowest;
                var fastest;
                for (var i in statistics.rts) {
                        tot = tot + statistics.rts[i];
                        if (statistics.rts[i] >= slowest || typeof slowest === "undefined") {
                                slowest = statistics.rts[i];
                        }
                        if (statistics.rts[i] <= fastest || typeof fastest == "undefined") {
                                fastest = statistics.rts[i];
                        }
                }
                return {Mean: Number(tot/statistics.NumRequests()).toFixed(3), Fastest: Number(fastest).toFixed(3), Slowest: Number(slowest).toFixed(3)}

          }
  }
  var ctx = $("#rtchart");

  var data = {
    labels: [], // empty for now
    datasets: [
      {
        label: "HTTP response times",
        backgroundColor: "#36A2EB",
        data: [] // empty for now
      }
    ]
  };

  var rtchart = new Chart(ctx, {
    type: 'line',
    data: data,
    fill: true,
  });

  var ctx2 = $("#tracechart");

  var tracedata = {
    labels: [], // empty for now
    datasets: [
      {
        label: "Dns lookup",
        backgroundColor: "#36A2EB",
        fill: false,
        data: [] // empty for now
      },
      {
        label: "Connection",
        backgroundColor: "#FF6384",
        fill: false,
        data: [] // empty for now
     },
     {
       label: "Request write",
       backgroundColor: "#FFCE56",
       fill: false,
       data: [] // empty for now
    },
    {
      label: "Request to first byte",
      backgroundColor: "#EC30E4",
      fill: false,
      data: [] // empty for now
   },
   {
     label: "Read response time",
     backgroundColor: "#61D512",
     fill: false,
     data: [] // empty for now
  },
    ]
  };

  var tracechart = new Chart(ctx2, {
    type: 'line',
    data: tracedata,
    fill: true,
  });

  var ctx3 = $("#codechart");
  var codedata = {
          labels: [], // empty for now
          datasets: [
                  {
                  data: [], // empty for now
                  backgroundColor: [
                  "#FF6384",
                  "#36A2EB",
                  "#FFCE56"
                  ],
                  hoverBackgroundColor: [
                          "#FF6384",
                          "#36A2EB",
                          "#FFCE56"
                  ]
                }
          ]
  };
  var codechart = new Chart(ctx3,{
      type: 'pie',
      data: codedata,
  });

  function receiveUpdate(obj) {
          statistics.rts.push(obj.Time);
          var stats = statistics.Calculated();

          var style = "black"
          if (obj.StatusCode >= 399) {
                  style = "red";
          }
          if (obj.Error != "") {
                  style = "red";
                  obj.StatusCode = obj.Error;
          }

          $('#messages').prepend("<li><span style='color:" + style + "'>HTTP Response seq=" + obj.Seq +  " status=" +  obj.StatusCode + " time=" + Number(obj.Time).toFixed(3) + "</span></li>" );
          $('#statistics').html("");
          $('#statistics').html("Testing: " + statistics.url + " Mean: " + stats.Mean + " Slowest: " + stats.Slowest + " Fastest: " + stats.Fastest);

          // Update rtgraph
          var d = new Date();
          var label = d.getHours() + ":" + d.getMinutes() + ":" + d.getSeconds()
          data.labels.push(label);
          data.datasets[0].data.push(Number(obj.Time).toFixed(3));
          rtchart.update();

          tracedata.labels.push(label);
          tracedata.datasets[0].data.push(Number(obj.DNSLookupTime).toFixed(5));
          tracedata.datasets[1].data.push(Number(obj.ConnectionTime).toFixed(5));
          tracedata.datasets[2].data.push(Number(obj.WriteRequestTime).toFixed(5));
          tracedata.datasets[3].data.push(Number(obj.WrittenToFirstByteTime).toFixed(5));
          tracedata.datasets[4].data.push(Number(obj.ReadResponseTime).toFixed(5));
          tracechart.update();

          // Update status code chart
          if (typeof statistics.statusCodes[obj.StatusCode] === "undefined") {
                  statistics.statusCodes[obj.StatusCode] = 1;
          } else {
                  statistics.statusCodes[obj.StatusCode] = statistics.statusCodes[obj.StatusCode] + 1;
          }

          codedata.labels = [];
          codedata.datasets[0].data = [];
          codes = statistics.Codes();
          for (var key in codes) {
                  codedata.labels.push(key);
                  codedata.datasets[0].data.push(codes[key]) ;
          }
          codechart.update();
  }

  $('#stopbutton').on('click', function(e) {
          $.ajax({
                  type: "POST",
                  url: "/target",
                  dataType: 'json',
                  async: false,
                  data: JSON.stringify({"stop": true}),
                  success: function () {
                          console.log("Yes");
                }
          })
          // Hide stop button
          $('#stopbutton').css("visibility", "hidden");

          // Hide spinner
          $('#spinner').css("visibility", "hidden");

          // Show start button
          $('#startbutton').css("visibility", "visible");
  })

  $('#goButton').on('click', function (e) {
          // Clear statistics
          statistics.statusCodes = {};
          statistics.errors = {};
          statistics.rts = [];
          statistics.url = $('#inputAddress').val();

          // Clear graphs
          data.labels = [];
          data.datasets[0].data = [];
          rtchart.update();

          tracedata.labels = [];
          tracedata.datasets[0].data = [];
          tracedata.datasets[1].data = [];
          tracedata.datasets[2].data = [];
          tracedata.datasets[3].data = [];
          tracedata.datasets[4].data = [];
          tracechart.update();

          codedata.labels = [];
          codedata.datasets[0].data = [];
          codechart.update();

          // Clear messages
          $('#messages').html("");

          // Send a post to start the actual job
          $.ajax({
                  type: "POST",
                  url: "/target",
                  dataType: 'json',
                  async: false,
                  data: JSON.stringify({
                          url: $('#inputAddress').val(),
                          sleepreq: parseInt($('#inputReqSleep').val()),
                          timeout: parseInt($('#inputTimeout').val())
                  }),
                  success: function () {
                          console.log("Yes");
                }
          })

          // Hide modal
          $('#newping').modal('hide');

          // Hide start button
          $('#startbutton').css("visibility", "hidden");

          // Show spinner
          $('#spinner').css("visibility", "visible");

          // Show stop button
          $('#stopbutton').css("visibility", "visible");

  });

  var loc = window.location;
  var ws_url = "ws://" + loc.host + "/ws";
  var ws = new WebSocket(ws_url);
  ws.onmessage = function(e) {
    var obj = JSON.parse(e.data);
    receiveUpdate(obj)
  };

  // Does not work
  $(window).on('beforeunload', function(){
    ws.close();
  });

});
