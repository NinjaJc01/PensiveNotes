function onLoginLoad() {
    const token = Cookies.get("SessionToken");
    if (token !== undefined && token !== "") {
        window.location = "/mynotes"
    }
    document.querySelector("#loginForm").addEventListener("submit", function (event) {
        //on pressing enter
        event.preventDefault()
        login()
    });
}

async function login() {
    const usernameBox = document.querySelector("#username");
    const passwordBox = document.querySelector("#password");
    const loginStatus = document.querySelector("#loginStatus");
    loginStatus.textContent = ""
    const creds = { username: usernameBox.value, password: passwordBox.value }
    const response = await postData("/api/user/login", creds)
    const statusOrCookie = await response.text()
    if (statusOrCookie === "") {
        return;
    }
    if (statusOrCookie === "Incorrect credentials" || statusOrCookie === "An unspecified error occurred") {
        loginStatus.textContent = "Incorrect Credentials"
        passwordBox.value=""
    } else {
        Cookies.set("SessionToken",statusOrCookie)
        console.log(statusOrCookie)
        window.location = "/mynotes"
    }
}