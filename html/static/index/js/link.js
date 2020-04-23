/* Makes links only open on left click and prevents dragging */
/*                    Why? 'cause why not                    */

var links = document.getElementsByTagName("A");

for (var i = 0; i < links.length; i++) {
    links[i].addEventListener("contextmenu", e => e.preventDefault());
    links[i].setAttribute('draggable', false);
}
