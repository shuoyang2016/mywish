function component() {
  var element = document.createElement('div');

  // Lodash, now imported by this script
  element.innerHTML = "<p>hello world!</p>"

  return element;
}

document.body.appendChild(component());
