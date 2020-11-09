package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	username = "root"
	password = "admin"
	hostname = "127.0.0.1:3306"
	dbname   = "goTest"
)

// Employee : fields
type Employee struct {
	ID     int     `json:"Id"`
	Name   string  `json:"Name"`
	Salary float64 `json:"Salary"`
}

var (
	employees []Employee
	conn      *sql.DB
)

func main() {
	conn = getDBConnection()
	defer conn.Close()
	pingDB(conn)
	handleRequests()
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true) //if end with slice then only get result
	router.HandleFunc("/", homePage)
	router.HandleFunc("/employees", allEmployees)                        //Read
	router.HandleFunc("/newEmployee", createNewEmployee).Methods("POST") //Create

	router.HandleFunc("/employees/{id}", getEmployee)                                   //Read
	router.HandleFunc("/updateEmployee", updateEmployee).Methods("PUT")                 //Update
	router.HandleFunc("/deleteEmployee", deleteEmployee).Methods("DELETE")              //Delete
	router.HandleFunc("/deleteEmployeeById/{id}", deleteEmployeeByID).Methods("DELETE") //Delete

	http.ListenAndServe(":10000", router)
}

func homePage(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: Home page!")
	fmt.Fprintf(w, "Welcome to Home page.")
}

func allEmployees(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: AllEmployees page!")
	employees = getEmployeesFromDB()
	json.NewEncoder(w).Encode(employees)
}

func createNewEmployee(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: creating new employee data page!")
	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	validateError(err)

	var emp Employee
	json.Unmarshal(reqBody, &emp)
	log.Println("Unmarshal emp:", emp)
	go insertIntoEmployeeTableDB(emp)
	json.NewEncoder(w).Encode(emp)
}

func getEmployee(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: get employee page!")
	vars := mux.Vars(r)
	employee := getEmployeebyIDFromDB(vars["id"])
	if employee.ID != 0 {
		json.NewEncoder(w).Encode(employee)
	} else {
		fmt.Fprintf(w, "Employee id not found!")
	}
}
func deleteEmployeeByID(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: Delete employee by ID page!")
	vars := mux.Vars(r)
	if deleteEmployeebyIDFromDB(vars["id"]) != 0 {
		fmt.Fprintf(w, "Deleted Employee id:"+vars["id"])
	} else {
		fmt.Fprintf(w, "Wrong Employee id!")
	}
}

func updateEmployee(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: update employee data page!")
	reqBody, _ := ioutil.ReadAll(r.Body)

	var emp Employee
	json.Unmarshal(reqBody, &emp)
	go updateIntoEmployeeTableDB(emp)
	json.NewEncoder(w).Encode(emp)
}

func deleteEmployee(w http.ResponseWriter, r *http.Request) {
	log.Println("Hitting end-point: delete employee data page!")
	reqBody, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	validateError(err)

	var emp Employee
	json.Unmarshal(reqBody, &emp)
	go deleteEmployeebyIDFromDB(emp.ID)
	go refreshPage(w)
}

func refreshPage(w http.ResponseWriter) {
	employees := getEmployeesFromDB()
	json.NewEncoder(w).Encode(employees)
}

//util
func validateError(err error) {
	if err != nil {
		panic(err)
	}
}

//DB process
func getDBConnection() *sql.DB {
	con, err := sql.Open("mysql", dsn())
	validateError(err)
	fmt.Println("DB connection successful!")
	con.Ping()
	return con
}
func dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbname)
}
func pingDB(con *sql.DB) {
	err := con.Ping()
	validateError(err)
}
func getEmployeesFromDB() []Employee {
	results, err := conn.Query("select * from Employee")
	validateError(err)
	employees = nil
	for results.Next() {
		var emp Employee
		err = results.Scan(&emp.ID, &emp.Name, &emp.Salary)
		validateError(err)
		employees = append(employees, emp)
	}
	return employees
}
func getEmployeebyIDFromDB(empID string) Employee {
	row := conn.QueryRow("SELECT id,Name,Salary FROM Employee where id = ?", empID)
	var emp Employee
	row.Scan(&emp.ID, &emp.Name, &emp.Salary)
	return emp
}
func deleteEmployeebyIDFromDB(empID interface{}) int64 {
	stmt, err := conn.Prepare("Delete From Employee where id = ?")
	validateError(err)
	res, err := stmt.Exec(empID)
	validateError(err)
	id, err := res.RowsAffected()
	validateError(err)
	if id == 0 {
		log.Printf("Not able to delete!")
	} else {
		log.Printf("Employee deleted successfully!")
	}

	return id
}

func insertIntoEmployeeTableDB(emp Employee) {
	stmt, err := conn.Prepare("Insert into Employee values(?,?,?)")
	validateError(err)
	_, err = stmt.Exec(emp.ID, emp.Name, emp.Salary)
	validateError(err)
}
func updateIntoEmployeeTableDB(emp Employee) {
	stmt, err := conn.Prepare("UPDATE Employee SET name=?, Salary=? where id=?")
	validateError(err)
	_, err = stmt.Exec(emp.Name, emp.Salary, emp.ID)
	validateError(err)
}
