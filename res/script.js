function toggleMobileMenu(menu) {
    menu.classList.toggle("open");
}

//Replace the whole page body but the tabs with the loading animation
document.addEventListener('click', function (event) {
	
	var loaderBody = `<div class="loader-container">
<div class="loader" scale=1></div>
<div>`
	if (["A", "BUTTON"].includes(event.target.tagName)) {
		//Select everything that isn't the navigation and remove it
		var divs = document.querySelectorAll("body > *:not(nav)");
		for(i=0;i<divs.length;i++) {
		  divs[i].remove()
		}
		// Add the loader div the body of the HTML that's left
		document.body.innerHTML += loaderBody
	}

}, false);
