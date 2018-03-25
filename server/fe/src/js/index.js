import '../css/cover.css';
import '../css/form.css';
import '../css/general.css';
import '../css/user_orders.css';
import '../css/create_product.css';
import '../css/index.css';
import '../css/debug.css';
import $ from 'jquery';


$('#recipeCarousel').carousel({
  interval: 10000
})

$('.carousel .carousel-item').each(function(){
    console.info("This is called.")
    var next = $(this).next();
    if (!next.length) {
        next = $(this).siblings(':first');
    }
    next.children(':first-child').clone().appendTo($(this));


    for (var i=0;i<2;i++) {
        next=next.next();
        if (!next.length) {
        	next = $(this).siblings(':first');
      	}
        next.children(':first-child').clone().appendTo($(this));
      }
});
