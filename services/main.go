package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type MealsRepository struct {
	db *sql.DB
}

type MealLog []Meal
type Meal struct {
	ID                   int
	Name                 string
	Time                 time.Time
	Calories             int64
	Protein, Fat, Carbs  int //grams
	VitaminA, VitaminB12 int //grams
}

const file string = "meallog.db"
const create string = `
  CREATE TABLE IF NOT EXISTS meallog (
  id INTEGER NOT NULL PRIMARY KEY,
  name TEXT,
  calories INTEGER,
  protein INTEGER,
  fat INTEGER,
  carbs INTEGER,
  vitaminA INTEGER,
  vitaminB INTEGER,
  time DATETIME NOT NULL
    );`

func NewMealLog() (*MealsRepository, error) {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(create); err != nil {
		return nil, err
	}
	return &MealsRepository{
		db: db,
	}, nil
}

func (c *MealsRepository) Insert(m Meal) (int, error) {
	res, err := c.db.Exec("INSERT INTO meallog VALUES(NULL,?,?,?,?,?,?,?,?);",
		m.Name,
		m.Calories,
		m.Protein,
		m.Fat,
		m.Carbs,
		m.VitaminA,
		m.VitaminB12,
		m.Time,
	)
	if err != nil {
		return 0, err
	}

	var id int64
	if id, err = res.LastInsertId(); err != nil {
		return 0, err
	}
	return int(id), nil
}

var mr, dberr = NewMealLog()

func getMeal(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/getMeal request\n")
	id_string := r.Header.Get("mealid")
	id, err := strconv.Atoi(id_string)
	if err != nil {
		io.WriteString(w, "err: provide mealid header key/val")
		return
	}
	row := mr.db.QueryRow("SELECT * FROM meallog WHERE id=?", id)
	mret := Meal{ID: id}
	err = row.Scan(&mret.ID, &mret.Name, &mret.Calories, &mret.Protein, &mret.Fat, &mret.Carbs, &mret.VitaminA, &mret.VitaminB12, &mret.Time)
	if err != nil {
		panic(err)
	}
	b, _ := json.Marshal(mret)
	io.WriteString(w, string(b)+"\n")
}

func getMeals(w http.ResponseWriter, r *http.Request) {
	meals := make([]Meal, 0)
	var mret Meal
	fmt.Printf("/getmeals request\n")
	rows, err := mr.db.Query("SELECT * FROM meallog")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&mret.ID, &mret.Name, &mret.Calories, &mret.Protein, &mret.Fat, &mret.Carbs, &mret.VitaminA, &mret.VitaminB12, &mret.Time)
		if err != nil {
			panic(err)
		}
		meals = append(meals, mret)

	}

	b, _ := json.Marshal(meals)
	io.WriteString(w, string(b)+"\n")
}

func addMeal(w http.ResponseWriter, r *http.Request) {
	var m = &Meal{}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, m)
	if err != nil {
		panic(err)
	}
	id, err := mr.Insert(*m)
	if err != nil {
		panic(err)
	}
	st := fmt.Sprintf("inserted meal! id = %d\n", id)
	io.WriteString(w, st)

}

func deleteMeal(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/deleteMeal request\n")
	id_string := r.Header.Get("mealid")
	id, err := strconv.Atoi(id_string)
	if err != nil {
		io.WriteString(w, "err: provide mealid header key/val")
		return
	}
	_, err = mr.db.Exec("DELETE FROM meallog WHERE id=?", id)

	if err == nil {
		io.WriteString(w, "success\n")
		return
	}
	io.WriteString(w, fmt.Sprintf("error:%s\n", err))
}

func main() {

	// var m Meal = Meal{
	// 	Calories:   800,
	// 	Protein:    27,
	// 	Fat:        20,
	// 	Carbs:      0,
	// 	VitaminA:   0,
	// 	VitaminB12: 0,
	// 	Time:       time.Now(),
	// }

	// id, err := mr.Insert(m)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("inserted meal! id = %d\n", id)

	// row := mr.db.QueryRow("SELECT * FROM meallog WHERE id=?", id)

	// mret := Meal{ID: id}
	// err = row.Scan(&mret.ID, &mret.Name, &mret.Calories, &mret.Protein, &mret.Fat, &mret.Carbs, &mret.VitaminA, &mret.VitaminB12, &mret.Time)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%#v\n", mret)
	http.HandleFunc("/getMeal", getMeal)
	http.HandleFunc("/getMeals", getMeals)
	http.HandleFunc("/addMeal", addMeal)
	http.HandleFunc("/deleteMeal", deleteMeal)

	http.ListenAndServe(":3333", nil)
}
