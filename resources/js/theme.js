
function setDark() {
    window.sessionStorage.setItem("theme", "dark")
    document.body.classList.remove("light");
    document.querySelector("nav").classList.remove("light");
    document.body.classList.add("dark");
    document.querySelector("nav").classList.add("dark");
    document.querySelectorAll(".subtitle").forEach(element => {
        element.classList.remove("light");
        element.classList.add("dark");
    });
    document.querySelectorAll("form").forEach(element => {
        if (element !== undefined && element !== null) {
            element.classList.remove("light");
            element.classList.add("dark");
        }
    });
}

function themeSwap() {
    const theme = window.sessionStorage.getItem("theme")
    if (theme === "light") {
        setDark()
        return
    }
    setLight()
}

function themeOnload() {
    const theme = window.sessionStorage.getItem("theme")
    if (theme === "light") {
        setLight()
        return
    }
    setDark()
}

function setLight() {
    window.sessionStorage.setItem("theme", "light");
    document.body.classList.remove("dark");
    document.querySelector("nav").classList.remove("dark");
    document.body.classList.add("light");
    document.querySelector("nav").classList.add("light");
    document.querySelectorAll(".subtitle").forEach(element => {
        element.classList.remove("dark");
        element.classList.add("light");
    });
    document.querySelectorAll("form").forEach(element => {
        if (element !== undefined && element !== null) {
            element.classList.remove("dark");
            element.classList.add("light");
        }
    });
}