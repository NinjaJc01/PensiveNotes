function addNote(note) {
    if (note !== undefined && note.noteTitle !== undefined && note.noteContent !== undefined) {
        //add it to the page
        const noteDiv = document.createElement("div");
        console.log(note)
        noteDiv.setAttribute("id", "note" + note.noteID);
        const noteTitle = document.createElement("h3");
        noteTitle.textContent = note.noteTitle;
        const noteContent = document.createElement("pre");
        const rule = document.createElement("hr");
        noteContent.textContent = note.noteContent;
        noteDiv.appendChild(noteTitle);
        noteDiv.appendChild(noteContent);
        noteDiv.appendChild(rule);
        document.querySelector("#notesDiv").appendChild(noteDiv);
    }
}

function logout() {
    postJsonData("/api/user/logout", "");
    window.location = "/";
}

async function postJsonData(url = '', data = {}) {
    // Default options are marked with *
    const response = await fetch(url, {
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            'Content-Type': 'application/json'
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *client
        body: JSON.stringify(data) // body data type must match "Content-Type" header
    });
    return response; // We don't always want JSON back
}

async function createNote() {
    const title = document.querySelector("#Title").value;
    const content = document.querySelector("#Content").value;
    console.log(title, content)
    postJsonData("/api/note/new", { noteTitle: title, noteContent: content }).then(function () {
        document.querySelector("#notesDiv").innerHTML = "";
        getNotes();
    })
    document.querySelector("#Title").value = "";
    document.querySelector("#Content").value = "";
    //delete notes, reload them
    //document.querySelector("#notesDiv").innerHTML = "";
    //window.setTimeout(await getNotes(), 200);
    cancelNote();
}

function noteDialog() {
    document.querySelector("#notesForm").classList.add("visible")
    document.querySelector("#notesForm").classList.remove("invisible")
    document.querySelector("#newNote").classList.add("invisible")
    document.querySelector("#newNote").classList.remove("visible")
}

function cancelNote() {
    document.querySelector("#notesForm").childNodes.forEach(element => {
        element.value = "";
    });
    document.querySelector("#notesForm").classList.add("invisible");
    document.querySelector("#notesForm").classList.remove("visible");
    document.querySelector("#newNote").classList.add("visible");
    document.querySelector("#newNote").classList.remove("invisible");
}
function onload() {
    const token = Cookies.get("SessionToken");
    if (token === undefined || token === "") {
        window.location = "/"
    }
}

async function getNotes() {
    const response = await getData("/api/note/list");
    if (response === null) {
        return;
    }
    response.sort((a, b) => (a.noteID > b.noteID) ? -1 : 1)
    response.forEach(element => {
        addNote(element);
    });
}