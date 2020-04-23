var startTime = new Date().getTime();

/* this JS handles the binary clock */

var main = function() {

    window.setInterval(function() {clock()}, 1000);
  
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
  
      var currentTime = new Date(new Date().getTime() - startTime)

      var hours = currentTime.getHours()-1;
      var minutes = currentTime.getMinutes();
      var seconds = currentTime.getSeconds();
  
      var s2 = seconds % 10;
      var s1 = (seconds - s2) / 10 % 10;
      var m2 = minutes % 10;
      var m1 = (minutes - m2) / 10 % 10;
      var h2 = hours % 10;
      var h1 = (hours - h2) / 10 % 10;
  
      lighting(s2, 's-2');
      lighting(s1, 's-1');
      lighting(m2, 'm-2');
      lighting(m1, 'm-1');
      lighting(h2, 'h-2');
      lighting(h1, 'h-1');
  
      $('.dec-hour1 p').text(h1);
      $('.dec-hour2 p').text(h2);
      $('.dec-minute1 p').text(m1);
      $('.dec-minute2 p').text(m2);
      $('.dec-second1 p').text(s1);
      $('.dec-second2 p').text(s2);
  
    };
  };

  $(document).ready(main);