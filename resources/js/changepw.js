async function login() {
    const usernameBox = document.querySelector("#username");
    const passwordBox = document.querySelector("#oldPassword");
    //const newPasswordBox1 = document.querySelector("#newPassword1");
    const newPasswordBox2 = document.querySelector("#newPassword2");
    const changeStatus = document.querySelector("#changePwStatus");
    changeStatus.textContent = "";
    const creds = {
        username: usernameBox.value,
        oldPassword: passwordBox.value,
        newPassword: newPasswordBox2.value
    };
    const response = await postData("/api/user/changepw", creds);
    const statusOrCookie = await response.text();
    if (statusOrCookie === "") {
        return;
    }
    if (statusOrCookie === "Incorrect credentials" || statusOrCookie === "An unspecified error occurred" || statusOrCookie === "Incomplete data") {
        changeStatus.textContent = "Incorrect Credentials";
        passwordBox.value = "";
    } else {
        changeStatus.textContent = "Successfully Updated Password";
        window.setTimeout(function () {
            logout();
        },300);
    }
}

function onPageLoad() {
    const newPasswordBox1 = document.querySelector("#newPassword1");
    const newPasswordBox2 = document.querySelector("#newPassword2");
    document.querySelector("#loginForm").addEventListener("submit", function (event) {
        //on pressing enter
        event.preventDefault()
        login()
    });
    newPasswordBox2.addEventListener("input", function (event) {
        if (newPasswordBox2.value !== newPasswordBox1.value) {
            newPasswordBox2.setCustomValidity("Passwords do not match");
        } else {
            newPasswordBox2.setCustomValidity("");
        }
    });
}