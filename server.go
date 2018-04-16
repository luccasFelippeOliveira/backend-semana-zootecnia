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

type Participante struct {
	ID          int    `json: "id"`
	CpfRa       int    `json: "cpf_ra"`
	CursoID     int    `json: "curso_id"`
	Nome        string `json: "nome"`
	Instituicao string `json: "instituicao"`
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

func dbGetMinicurso(dataInicio *int, dataFim *int) []Minicurso {
	var rows *sql.Rows
	var err error

	if dataInicio == nil {
		rows, err = db.Query("SELECT * FROM minicurso")
	} else {
		rows, err =
			db.Query("SELECT * FROM minicurso"+
				" where horario_comeco > datetime(?, \"unixepoch\", \"localtime\")"+
				" AND horario_fim < datetime(?, \"unixepoch\", \"localtime\")"+
				" AND vagas_restantes > 0", *dataInicio, *dataFim)
	}

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
	minicursos := dbGetMinicurso(nil, nil)
	if minicursos != nil {
		return c.JSON(http.StatusOK, minicursos)
	}
	return fmt.Errorf("Minicurso is null")
}

func GetMiniCursosData(c echo.Context) error {
	// Pega horario maximo.
	body := &struct {
		DataInicio int
		DataFim    int
	}{}

	if err := c.Bind(body); err != nil {
		return err
	}

	minicursos := dbGetMinicurso(&body.DataInicio, &body.DataFim)
	if minicursos != nil {
		return c.JSON(http.StatusOK, minicursos)
	}

	return fmt.Errorf("minicursos is null")
}

func PostNovaInscricao(c echo.Context) error {

	return nil
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
	e.POST("/minicursos", GetMiniCursosData)
	e.POST("/inscricao", PostNovaInscricao)

	// Start Server
	e.Logger.Fatal(e.Start(":1323"))
}
