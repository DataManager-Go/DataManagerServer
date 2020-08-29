// Prevent text marking on this page
onselectstart = (e) => {e.preventDefault()}

// Start date
var d = new Date();
// Covert current time to UTC
var startTime = d.getTime() + d.getTimezoneOffset() * 60000;

// Timer
var timer;

var main = function() {

    // Update everys 1000 ms
    timer = window.setInterval(function() {clock()}, 1000);
  
    // Pre-Setup to make everything nice and dark
    var lighting = function(n,t) {
      if(n == 1) {
        $('.'+t+'-1').addClass('on').removeClass('off');
        $('.'+t+'-2').addClass('off').removeClass('on');
        $('.'+t+'-4').addClass('off').removeClass('on');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 2) {
        $('.'+t+'-1').addClass('off').removeClass('on');
        $('.'+t+'-2').addClass('on').removeClass('off');
        $('.'+t+'-4').addClass('off').removeClass('on');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 3) {
        $('.'+t+'-1').addClass('on').removeClass('off');
        $('.'+t+'-2').addClass('on').removeClass('off');
        $('.'+t+'-4').addClass('off').removeClass('on');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 4) {
        $('.'+t+'-1').addClass('off').removeClass('on');
        $('.'+t+'-2').addClass('off').removeClass('on');
        $('.'+t+'-4').addClass('on').removeClass('off');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 5) {
        $('.'+t+'-1').addClass('on').removeClass('off');
        $('.'+t+'-2').addClass('off').removeClass('on');
        $('.'+t+'-4').addClass('on').removeClass('off');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 6) {
        $('.'+t+'-1').addClass('off').removeClass('on');
        $('.'+t+'-2').addClass('on').removeClass('off');
        $('.'+t+'-4').addClass('on').removeClass('off');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 7) {
        $('.'+t+'-1').addClass('on').removeClass('off');
        $('.'+t+'-2').addClass('on').removeClass('off');
        $('.'+t+'-4').addClass('on').removeClass('off');
        $('.'+t+'-8').addClass('off').removeClass('on');
  
      } else if(n == 8) {
        $('.'+t+'-1').addClass('off').removeClass('on');
        $('.'+t+'-2').addClass('off').removeClass('on');
        $('.'+t+'-4').addClass('off').removeClass('on');
        $('.'+t+'-8').addClass('on').removeClass('off');
  
      } else if(n == 9) {
        $('.'+t+'-1').addClass('on').removeClass('off');
        $('.'+t+'-2').addClass('off').removeClass('on');
        $('.'+t+'-4').addClass('off').removeClass('on');
        $('.'+t+'-8').addClass('on').removeClass('off');

      } else {
        $('.'+t+'-1').addClass('off').removeClass('on');
        $('.'+t+'-2').addClass('off').removeClass('on');
        $('.'+t+'-4').addClass('off').removeClass('on');
        $('.'+t+'-8').addClass('off').removeClass('on');

      };
    };

    var clock = function() {

      // Current Time
      var d2 = new Date();
      var currentTime = d2.getTime() + (d2.getTimezoneOffset() * 60000) - startTime;

      // Convert miliseconds
      var seconds = currentTime / 1000;
      var minutes = seconds / 60;
      var hours = minutes / 60;

      // Get within current limits
      seconds = Math.floor(seconds % 60);
      minutes = Math.floor(minutes % 60);
      hours = Math.floor(hours);

      // Negative time protection
      if (hours < 0 || minutes < 0 || seconds < 0) {
        d = new Date();
        startTime = d.getTime() + d.getTimezoneOffset() * 60000;
        return;
      }

      // No-Life protection
      if (hours >= 30) {
        document.getElementById("clockDIV").innerHTML = "";
        document.getElementById("numberDIV").innerHTML = "";

        // Get some help.
        var timeMsg = document.createElement("h1");
        timeMsg.innerHTML = " > 30 GODDAMNED HOURS!";

        var helpMsg = document.createElement("h3");
        helpMsg.innerHTML = "If you didn't cheat on achieving this amount of time<br> you should really go out and get some help."
        helpMsg.style.textAlign = "center";
        helpMsg.style.color = "#4f4f66"

        document.getElementById("innerDIV").appendChild(timeMsg);
        document.getElementById("innerDIV").appendChild(helpMsg);

        window.clearInterval(timer);
      } 

      // Calculate the binary lights
      var s2 = seconds % 10;
      var s1 = (seconds - s2) / 10 % 10;
      var m2 = minutes % 10;
      var m1 = (minutes - m2) / 10 % 10;
      var h2 = hours % 10;
      var h1 = (hours - h2) / 10 % 10;
  
      // Make it bright (or dark lol)
      lighting(s2, 's-2');
      lighting(s1, 's-1');
      lighting(m2, 'm-2');
      lighting(m1, 'm-1');
      lighting(h2, 'h-2');
      lighting(h1, 'h-1');
  
      // Display time number
      $('.dec-hour1 p').text(h1);
      $('.dec-hour2 p').text(h2);
      $('.dec-minute1 p').text(m1);
      $('.dec-minute2 p').text(m2);
      $('.dec-second1 p').text(s1);
      $('.dec-second2 p').text(s2);
  
    };
  };

  $(document).ready(main);