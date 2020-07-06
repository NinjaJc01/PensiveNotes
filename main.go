package main
//TODO
/*
- Implement Password Change Handler
- Write frontend
- Create database to ship with PensiveNotes
*/
import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

//User represents a user entity
type User struct {
	UserID       int
	Username     string
	PasswordHash string
	Salt         string
	SessionToken string
}

//UserReq is the format of a user creation request
type UserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//UserReqPWChange is the format of a password change request
type UserReqPWChange struct {
	Username string `json:"username"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword string `json:"newPassword"`
}

//Note represents the structure of a note in the database
type Note struct {
	NoteID  int    `json:"noteID"`
	UserID  int    `json:"userID"`
	Title   string `json:"noteTitle"`
	Content string `json:"noteContent"`
}

//NoteRequest is similar to Note, but includes a session token to verify the user for creation instead of uid
type NoteRequest struct {
	SessionToken string `json:"sessionToken"`
	Content      string `json:"noteContent"`
	Title        string `json:"noteTitle"`
}

var database *sql.DB

func main() {
	//Connect to the db
	databasePointer, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Database connection failed")
	}
	database = databasePointer
	startServer()
}

func generateToken() string {
	token := make([]byte, 16)
	rand.Read(token)
	return (fmt.Sprintf("%x", token)) //hex encoded token
}

func generateSalt() string {
	token := make([]byte, 16)
	rand.Read(token)
	return (fmt.Sprintf("%x", token)) //hex encoded salt
}

func startServer() {
	portPtr := flag.Int("p", 8081, "Port number to run the server on")
	flag.Parse()
	port := *portPtr
	mr := mux.NewRouter()
	mr.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	apiRouter := mr.PathPrefix("/api").Subrouter()
	go mime.AddExtensionType(".css", "text/css; charset=utf-8")
	go mime.AddExtensionType(".js", "application/javascript; charset=utf-8")
	//Setup a static router for HTML/CSS/JS
	mr.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./resources"))))
	//CRUD API routes
	userRouter := apiRouter.PathPrefix("/user").Subrouter()
	/*Create User	*/ userRouter.HandleFunc("/create", createUser).Methods("POST")
	/*Change passwd	*/ userRouter.HandleFunc("/changepw", doChangePassword).Methods("POST")
	/*Log out		*/ userRouter.HandleFunc("/logout", doLogout).Methods("POST")
	/*Log In		*/ userRouter.HandleFunc("/login", doLogin).Methods("POST")
	noteRouter := apiRouter.PathPrefix("/note").Subrouter()
	/*Create		*/ noteRouter.HandleFunc("/new", createNote).Methods("POST")
	/*List by user	*/ noteRouter.HandleFunc("/list", listNotesForUser).Methods("GET")
	/*Read one		*/ noteRouter.HandleFunc("/{id}", readNote).Methods("GET")
	log.Println("Listening for requests on",fmt.Sprintf(":%v", port))
	http.ListenAndServe(fmt.Sprintf(":%v", port), mr)
}

func verifyPass(user User, password string) bool {
	resultHash := hashPassword(password, user.Salt)
	return resultHash == user.PasswordHash
}

func hashPassword(password string, salt string) string {
	hash := sha512.Sum512([]byte(password + salt))
	return fmt.Sprintf("%x", hash)
}

func reqHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) { //Handle 404s
	w.WriteHeader(404)
}

func userReqToUser(req UserReq) User {
	var user User
	user.UserID = dbNextUserID()
	user.Username = req.Username
	user.Salt = generateSalt()
	user.PasswordHash = hashPassword(req.Password, user.Salt)
	return user
}

func doLogout(w http.ResponseWriter, r *http.Request) {
	//determine user by cookie
	token, err := r.Cookie("SessionToken")
	http.SetCookie(w, &http.Cookie{Name: "SessionToken", Value: "", Path: "/"})
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(401) //No cookie, no auth
		return
	}
	user, status := dbUserVerifyToken(token.Value)
	if status != "success" { //the user must be authenticated
		w.WriteHeader(401)
		return
	}
	status = dbUserSetToken("", user.UserID)
	if status != "success" {
		w.WriteHeader(500)
		return
	}
}

func doChangePassword(w http.ResponseWriter, r *http.Request) {
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	log.Println("Couldn't read body")
	// 	return
	// }
	// log.Println(string(body))
	username := r.FormValue("username")
	oldPassword := r.FormValue("oldPassword")
	newPassword := r.FormValue("newPassword")
	log.Println(username,oldPassword,newPassword)
	if username == "" || oldPassword == "" || newPassword == "" {
		w.Write([]byte("Incomplete data"))
		return
	}
	//Verify user and password
	user, status := dbUserGetByUsername(username)
	//on success, set cookie and send them to the flag
	if status != "success" {
		hashPassword("nottherealpassword", "aSalt")
		w.Write([]byte("Incorrect credentials"))
		return //error!
	}
	if !verifyPass(user, oldPassword) {
		w.Write([]byte("Incorrect credentials"))
		return //error!
	}
	user.Salt = generateSalt()
	user.PasswordHash = hashPassword(newPassword,user.Salt)
	status = dbUserSetPassword(user.PasswordHash, user.Salt, user.UserID)
	if status != "success" {
		w.Write([]byte("An unspecified error occurred"))
		log.Println(status)
		return
	}
	w.Write([]byte("Successfully changed your password."))
}

func doLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		w.WriteHeader(400)
		w.Write([]byte("Incorrect credentials"))
		return
	}
	//Verify user and password
	user, status := dbUserGetByUsername(username)
	//on success, set cookie and send them to the flag
	if status != "success" {
		hashPassword("nottherealpassword", "aSalt")
		w.Write([]byte("Incorrect credentials"))
		return //error!
	}
	if !verifyPass(user, password) {
		w.Write([]byte("Incorrect credentials"))
		return //error!
	}
	user.SessionToken = generateToken()
	status = dbUserSetToken(user.SessionToken, user.UserID)
	if status != "success" {
		w.Write([]byte("An unspecified error occurred"))
		return //error!
	}
	//give them a session token in return and TELL THEM it worked
	http.SetCookie(w, &http.Cookie{Name: "SessionToken", Value: user.SessionToken, Path: "/"})
	//http.Redirect(w, r, "/mynotes", 303)
	w.Write([]byte(user.SessionToken))
}

func createUser(w http.ResponseWriter, r *http.Request) { //
	response := make(map[string]string)
	var userReq UserReq
	//Try to read JSON in, make sure it's correct
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		response["status"] = "Error creating user"
		json.NewEncoder(w).Encode(response)
		return
	}
	err = json.Unmarshal(body, &userReq)
	if err != nil {
		w.WriteHeader(400)
		response["status"] = "Error parsing JSON"
		json.NewEncoder(w).Encode(response)
		log.Println(string(body), userReq)
		return
	}
	newUser := userReqToUser(userReq)
	status := dbUserCreate(newUser)
	if status != "success" {
		w.WriteHeader(500)
		response["status"] = "Error creating user"
		json.NewEncoder(w).Encode(response)
		return
	}
	response["status"] = "success"
	json.NewEncoder(w).Encode(response)
	return
}

func createNote(w http.ResponseWriter, r *http.Request) {
	/*
		{
			sessionToken:
			noteTitle:
			noteContent:
		}
	*/
	response := make(map[string]string)
	token, err := r.Cookie("SessionToken")
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(401) //No cookie, no auth
		response["status"] = "Unauthenticated"
		json.NewEncoder(w).Encode(response)
		return
	}
	user, status := dbUserVerifyToken(token.Value)
	if status != "success" { //the user must be authenticated
		w.WriteHeader(401)
		response["status"] = "Unauthenticated"
		json.NewEncoder(w).Encode(response)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		response["status"] = "Error creating note"
		json.NewEncoder(w).Encode(response)
		return
	}
	var note Note
	err = json.Unmarshal(body, &note)
	note.UserID = user.UserID
	note.NoteID = dbNoteNextID()
	dbNoteCreate(note)

	response["status"] = "Successfully created note"
	json.NewEncoder(w).Encode(response)
	return
}

func listNotesForUser(w http.ResponseWriter, r *http.Request) {
	//determine what user by token
	token, err := r.Cookie("SessionToken")
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(401) //No cookie, no auth
		return
	}
	user, status := dbUserVerifyToken(token.Value)
	if status != "success" {
		w.WriteHeader(401)
	}
	notes, status := dbNoteGetAllByUserID(user.UserID)
	if status != "success" {
		w.WriteHeader(500)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func readNote(w http.ResponseWriter, r *http.Request) { //Note by ID
	//check perms for the user, notes are private
	//then check noteid
	//then write it as JSON to w.
	return
}

func updateNote(w http.ResponseWriter, r *http.Request) {
	return
}

func deleteNote(w http.ResponseWriter, r *http.Request) {
	return
}

/* DATABASE */
/* User Database Functions */
func dbNextUserID() int {
	res, err := database.Query("SELECT Max(UserID) FROM User")
	if err != nil {
		if res != nil {
			res.Close()
		}
		log.Println(err.Error())
		return -1
	}
	var UserID int
	if res.Next() {
		err = res.Scan(&UserID)
		if err != nil {
			res.Close()
			log.Println(err.Error())
			return -1
		}
	} else {
		log.Println("No Next ID for User found?")
		return -1
	}
	res.Close()
	return UserID + 1
}

func dbUserCreate(user User) (status string) {
	statement, err := database.Prepare("INSERT INTO User(UserID, Username, HashedPassword, Salt, SessionToken) values(?,?,?,?,?)")
	if err != nil {
		log.Println(err.Error())
		return "DB statement preperation failed"
	}
	_, err = statement.Exec(
		user.UserID,
		user.Username,
		user.PasswordHash,
		user.Salt,
		user.SessionToken)
	if err != nil {
		log.Println(err.Error())
		return "DB statement execution failed"
	}
	return "success"
}

func dbUserSetToken(token string, userID int) string {
	statement, err := database.Prepare("UPDATE User SET SessionToken = ? WHERE UserID = ?")
	if err != nil {
		log.Println(err.Error())
		return "DB statement preperation failed"
	}
	_, err = statement.Exec(token, userID)
	if err != nil {
		log.Println(err.Error())
		return "DB execution failed"
	}

	return "success"
}

func dbUserSetPassword(newHash string, newSalt string, userID int) string {
	statement, err := database.Prepare("UPDATE User SET HashedPassword = ?, Salt = ? WHERE UserID = ?")
	if err != nil {
		log.Println(err.Error())
		return "DB statement preperation failed"
	}
	_, err = statement.Exec(newHash, newSalt, userID)
	if err != nil {
		log.Println(err.Error())
		return "DB execution failed"
	}

	return "success"
}

func dbUserVerifyToken(token string) (User, string) {
	//Lookup by Token
	if token == "" {
		return User{}, "nil token"
	}
	statement, err := database.Prepare("SELECT * FROM User WHERE SessionToken = ?")
	if err != nil {
		log.Println(err.Error())
		return User{}, "DB statement preperation failed"
	}
	res, err := statement.Query(token)
	if err != nil {
		res.Close()
		log.Println(err.Error())
		return User{}, "DB execution failed"
	}
	var user User
	for res.Next() { //Technically Identical UIDs will give the last
		var (
			UserID       int
			Username     string
			PasswordHash string
			Salt         string
			SessionToken string
		)

		err = res.Scan(&UserID, &Username, &PasswordHash, &Salt, &SessionToken)
		if err != nil {
			res.Close()
			log.Println(err.Error())
			return User{}, "DB query reading failed"
		}
		user = User{
			UserID:       UserID,
			Username:     Username,
			PasswordHash: PasswordHash,
			Salt:         Salt,
			SessionToken: SessionToken,
		}
	}
	if user.UserID == 0 {
		return User{}, "Invalid token"
	}
	res.Close()
	return user, "success"
}

func dbUserGetByUsername(username string) (User, string) {
	statement, err := database.Prepare("SELECT * FROM User WHERE Username = ?")
	if err != nil {
		log.Println(err.Error())
		return User{}, "DB statement preperation failed"
	}
	res, err := statement.Query(username)
	if err != nil {
		res.Close()
		log.Println(err.Error())
		return User{}, "DB execution failed"
	}
	var user User
	for res.Next() { //Technically Identical Usernames will give the last
		var (
			UserID       int
			Username     string
			PasswordHash string
			Salt         string
			SessionToken string
		)

		err = res.Scan(&UserID, &Username, &PasswordHash, &Salt, &SessionToken)
		if err != nil {
			res.Close()
			log.Println(err.Error())
			return User{}, "DB query reading failed"
		}
		user = User{
			UserID:       UserID,
			Username:     Username,
			PasswordHash: PasswordHash,
			Salt:         Salt,
			SessionToken: SessionToken,
		}
	}
	if user.UserID == 0 {
		return User{}, "User not found"
	}
	res.Close()
	return user, "success"
}

/*Note Database Functions*/
func dbNoteNextID() int {
	res, err := database.Query("SELECT Max(NoteID) FROM Note")
	if err != nil {
		log.Println(err.Error())
		return -1
	}
	var NoteID int
	if res.Next() {
		err = res.Scan(&NoteID)
		if err != nil {
			log.Println(err.Error())
			return -1
		}
	} else {
		res.Close()
		log.Println("No Next ID for Note found?")
		return -1
	}
	res.Close()
	return NoteID + 1
}

func dbNoteCreate(note Note) string {
	statement, err := database.Prepare(
		"INSERT INTO Note(NoteID, UserID, NoteTitle, NoteContent) values(?,?,?,?)")
	if err != nil {
		log.Println(err.Error())
		return "DB statement preperation failed"
	}
	_, err = statement.Exec(
		note.NoteID,
		note.UserID,
		note.Title,
		note.Content)
	if err != nil {
		log.Println(err.Error())
		return "DB statement execution failed"
	}
	return "success"
}

func dbNoteGetAll() ([]Note, string) {
	var noteList []Note
	res, err := database.Query("SELECT * FROM Note")
	if err != nil {
		log.Println(err.Error())
		return nil, "DB statement execution failed"
	}
	for res.Next() {
		var (
			NoteID  int
			UserID  int
			Title   string
			Content string
		)
		var note Note
		err = res.Scan(&NoteID, &UserID, &Title, &Content)
		if err != nil {
			res.Close()
			log.Println(err.Error())
			return nil, "DB statement execution failed"
		}
		note = Note{
			NoteID:  NoteID,
			UserID:  UserID,
			Title:   Title,
			Content: Content,
		}
		noteList = append(noteList, note)
	}
	res.Close()
	return noteList, "success"
}

func dbNoteGetByNoteID(id int) (Note, string) {
	statement, err := database.Prepare("SELECT * FROM Note WHERE NoteID = ?")
	if err != nil {
		log.Println(err.Error())
		return Note{}, "DB statement preperation failed"
	}
	res, err := statement.Query(id)
	if err != nil {
		log.Println(err.Error())
		return Note{}, "DB execution failed"
	}
	var note Note
	for res.Next() {
		var (
			NoteID  int
			UserID  int
			Title   string
			Content string
		)
		err = res.Scan(&NoteID, &UserID, &Title, &Content)
		if err != nil {
			log.Println(err.Error())
			res.Close()
			return Note{}, "DB statement execution failed"
		}
		note = Note{
			NoteID:  NoteID,
			UserID:  UserID,
			Title:   Title,
			Content: Content,
		}
	}
	res.Close()
	return note, "success"
}

func dbNoteGetAllByUserID(id int) ([]Note, string) {
	statement, err := database.Prepare("SELECT * FROM Note WHERE UserID = ?")
	if err != nil {
		log.Println(err.Error())
		return nil, "DB statement preperation failed"
	}
	res, err := statement.Query(id)
	if err != nil {
		log.Println(err.Error())
		return nil, "DB execution failed"
	}
	var notes []Note
	for res.Next() {
		var (
			note    Note
			NoteID  int
			UserID  int
			Title   string
			Content string
		)
		err = res.Scan(&NoteID, &UserID, &Title, &Content)
		if err != nil {
			log.Println(err.Error())
			res.Close()
			return nil, "DB statement execution failed"
		}
		note = Note{
			NoteID:  NoteID,
			UserID:  UserID,
			Title:   Title,
			Content: Content,
		}
		notes = append(notes, note)
	}
	res.Close()
	return notes, "success"
}
