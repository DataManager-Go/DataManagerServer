const themeMap = {
  dark: "light",
  light: "solar",
  solar: "dark"
};

const theme = localStorage.getItem('theme')
  || (tmp = Object.keys(themeMap)[0],
      localStorage.setItem('theme', tmp),
      tmp);
const bodyClass = document.body.classList;
bodyClass.add(theme);

function toggleTheme() {
  // Nav Bar
  const current = localStorage.getItem('theme');
  const next = themeMap[current];

  bodyClass.replace(current, next);
  localStorage.setItem('theme', next);

  // Body
  const currentBody = localStorage.getItem('theme');
  const nextBody = themeMap[currentBody];

  document.body.replace(currentBody, nextBody);
  localStorage.setItem('theme', nextBody);
}

document.getElementById('themeButton').onclick = toggleTheme;
