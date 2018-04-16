package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/mattn/go-sqlite3"
)

type Curso struct {
	ID   int    `json: "id"`
	Nome string `json: "nome"`
}

type Minicurso struct {
	ID             int       `json: "id"`
	Nome           string    `json: "nome"`
	Palestrante    string    `json: "palestrante"`
	HorarioComeco  time.Time `json: "horarioComeco"`
	HorarioFim     time.Time `json:"horarioFim"`
	Vagas          int       `json:"vagas"`
	VagasRestantes int       `json: "vagasRestantes"`
}

var (
	db *sql.DB
)

func dbGetCursos() []Curso {
	rows, err := db.Query("SELECT * FROM curso")
	if err != nil {
		return nil
	}

	cursos := make([]Curso, 0)
	defer rows.Close()
	for rows.Next() {
		var id int
		var nomeByte []byte
		err := rows.Scan(&id, &nomeByte)
		nome := string(nomeByte)
		if err != nil {
			return nil
		}
		cursos = append(cursos, Curso{ID: id, Nome: nome})
	}
	return cursos
}

func dbGetMinicurso() []Minicurso {
	rows, err := db.Query("SELECT * FROM minicurso")
	if err != nil {
		return nil
	}

	minicursos := make([]Minicurso, 0)
	defer rows.Close()
	for rows.Next() {
		var id int
		var nome string
		var palestrante string
		var horarioInicio time.Time
		var horarioFim time.Time
		var vagas int
		var vagasRestantes int

		err := rows.Scan(
			&id,
			&nome,
			&palestrante,
			&horarioInicio,
			&horarioFim,
			&vagas,
			&vagasRestantes)

		if err != nil {
			fmt.Println("Cannot rows.Scan")
			return nil
		}

		minicursos = append(minicursos,
			Minicurso{
				ID:             id,
				Nome:           nome,
				Palestrante:    palestrante,
				HorarioComeco:  horarioInicio,
				HorarioFim:     horarioFim,
				Vagas:          vagas,
				VagasRestantes: vagasRestantes})
	}
	return minicursos
}

func GetCursos(c echo.Context) error {
	cursos := dbGetCursos()
	if cursos != nil {
		//fmt.Println(cursos[1])
		return c.JSON(http.StatusOK, cursos)
	}
	return fmt.Errorf("cursos is null")
}

func GetMiniCursos(c echo.Context) error {
	minicursos := dbGetMinicurso()
	if minicursos != nil {
		return c.JSON(http.StatusOK, minicursos)
	}
	return fmt.Errorf("Minicurso is null")
}

func main() {
	// Echo instance
	e := echo.New()

	openDB, err := sql.Open("sqlite3", "./db/semana-zoo.db")
	if err != nil {
		e.Logger.Fatalf("Can't open databasa %e", err)
	}
	db = openDB

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Route -> Handler
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World\n")
	})

	// e.GET("/cursos", func(c echo.Context) error {
	// 	return c.JSON(http.StatusOK, cursos)
	// })

	e.GET("/cursos", GetCursos)
	e.GET("/minicursos", GetMiniCursos)

	// Start Server
	e.Logger.Fatal(e.Start(":1323"))
}
