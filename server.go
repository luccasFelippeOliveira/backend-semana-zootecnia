package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/mattn/go-sqlite3"
)

type Curso struct {
	ID   int    `json:"id"`
	Nome string `json:"nome"`
}

type Minicurso struct {
	ID              int       `json:"id"`
	Nome            string    `json:"nome"`
	Palestrante     string    `json:"palestrante"`
	HorarioComeco   time.Time `json:"horarioComeco"`
	HorarioFim      time.Time `json:"horarioFim"`
	Vagas           int       `json:"vagas"`
	VagasRestantes  int       `json:"vagasRestantes"`
	QuantidadeHoras int       `json:"quantidadeHoras"`
}

type Participante struct {
	ID          int    `json:"id"`
	CpfRa       int    `json:"cpf_ra"`
	CursoID     int    `json:"curso_id"`
	Nome        string `json:"nome"`
	Instituicao string `json:"instituicao"`
	Pago        bool   `json:"pago"`
}

type ParticpanteRequest struct {
	CpfRa       string      `json:"cpf_ra"`
	Nome        string      `json:"nome"`
	Instituicao string      `json:"instituicao"`
	Curso       Curso       `json:"curso"`
	Minicrusos  []Minicurso `json:"minicursos"`
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
	rows, err := db.Query("SELECT * FROM minicurso where vagas_restantes > 0")
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
		var quantidadeHoras int

		err := rows.Scan(
			&id,
			&nome,
			&palestrante,
			&horarioInicio,
			&horarioFim,
			&vagas,
			&vagasRestantes,
			&quantidadeHoras)

		if err != nil {
			fmt.Println("Cannot rows.Scan")
			return nil
		}

		minicursos = append(minicursos,
			Minicurso{
				ID:              id,
				Nome:            nome,
				Palestrante:     palestrante,
				HorarioComeco:   horarioInicio,
				HorarioFim:      horarioFim,
				Vagas:           vagas,
				VagasRestantes:  vagasRestantes,
				QuantidadeHoras: quantidadeHoras})
	}
	return minicursos
}

func dbInsereParticipante(p *Participante) (int64, error) {
	// Por default, assume-se que o aluno nao pagou sua participacao
	queryString := "INSERT INTO participante(cpf_ra, curso_id, nome, instituicao, pago) values (?, ?, ?, ?, 0)"

	stmt, err := db.Prepare(queryString)
	if err != nil {
		return -1, err
	}

	res, err := stmt.Exec(p.CpfRa, p.CursoID, p.Nome, p.Instituicao)
	if err != nil {
		return -1, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return id, err
}

func dbAtualizaParticipante(p *Participante) error {
	query := "UPDATE participante SET pago = ? where participante_id = ?"

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	pago := 0
	if p.Pago {
		pago = 1
	}

	_, err = tx.Stmt(stmt).Exec(pago, p.ID)
	if err != nil {
		fmt.Println("Rollback dbAtualizaParticipante")
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return nil
}

func dbAtualizaVaga(m *Minicurso) error {
	query := "UPDATE minicurso SET vagas_restantes = vagas_restantes - 1	WHERE minicurso_id = ?"

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(m.ID)
	if err != nil {
		fmt.Println("Doing rollback dbAtualizaVaga")
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return nil
}

func dbInscreverMinicurso(m *Minicurso, p *Participante) error {
	query := "INSERT INTO participante_minicurso(participante_id, minicurso_id) VALUES (?,?)"

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(p.ID, m.ID)
	if err != nil {
		fmt.Println("Doing rollback dbInscreverMinicurso")
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return nil
}

func dbRemoverInscricao(m *Minicurso, p *Participante) error {
	query := "DELETE FROM participante_minicurso WHERE participante_id = ? AND minicurso_id = ?"

	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Stmt(stmt).Exec(p.ID, m.ID)
	if err != nil {
		fmt.Println("Doing rollback dbRemoverInscricao")
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return nil
}

func dbGetParticipante(cpfRa int) (*Participante, error) {
	query := "SELECT * FROM participante WHERE cpf_ra = ?"

	rows, err := db.Query(query, cpfRa)
	if err != nil {
		return nil, err
	}

	participante := new(Participante)

	defer rows.Close()
	for rows.Next() {
		var ID int
		var CpfRa int
		var CursoID int
		var Nome string
		var Instituicao string
		var Pago bool

		err := rows.Scan(
			&ID,
			&CpfRa,
			&CursoID,
			&Nome,
			&Instituicao,
			&Pago)

		if err != nil {
			return nil, err
		}

		participante.ID = ID
		participante.CpfRa = CpfRa
		participante.CursoID = CursoID
		participante.Nome = Nome
		participante.Instituicao = Instituicao

		return participante, nil
	}

	return nil, nil
}

func dbGetParticipanteId(id int) (*Participante, error) {
	query := "SELECT * FROM participante WHERE participante_id = ?"

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}

	participante := new(Participante)

	defer rows.Close()
	for rows.Next() {
		var ID int
		var CpfRa int
		var CursoID int
		var Nome string
		var Instituicao string
		var Pago bool

		err := rows.Scan(
			&ID,
			&CpfRa,
			&CursoID,
			&Nome,
			&Instituicao,
			&Pago)

		if err != nil {
			return nil, err
		}

		participante.ID = ID
		participante.CpfRa = CpfRa
		participante.CursoID = CursoID
		participante.Nome = Nome
		participante.Instituicao = Instituicao

		return participante, nil
	}

	return nil, nil
}

func dbGetParticipantes() ([]Participante, error) {
	query := "SELECT * FROM participante"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	participantes := make([]Participante, 0)

	defer rows.Close()
	for rows.Next() {
		var iD int
		var cpfRa int
		var cursoID int
		var nome string
		var instituicao string
		var pago bool

		err := rows.Scan(
			&iD,
			&cpfRa,
			&cursoID,
			&nome,
			&instituicao,
			&pago)

		if err != nil {
			return nil, err
		}

		participantes = append(participantes,
			Participante{
				ID:          iD,
				CpfRa:       cpfRa,
				CursoID:     cursoID,
				Nome:        nome,
				Instituicao: instituicao,
				Pago:        pago})
	}
	return participantes, nil
}

func atualizaVagas(m *Minicurso, p *Participante) error {
	err := dbInscreverMinicurso(m, p)
	if err != nil {
		return fmt.Errorf("Failed to inscrever minicurso")
	}

	err = dbAtualizaVaga(m)
	if err != nil {
		// Remove inscricao
		if e := dbRemoverInscricao(m, p); e != nil {
			return fmt.Errorf("Failed to remove inscricao")
		}
		return fmt.Errorf("Failed to atualizaVagas")
	}
	return nil
}

func novaInscricao(inscricao *ParticpanteRequest) error {
	// Insere novo participante
	p := Participante{}
	p.Nome = inscricao.Nome

	if (inscricao.Curso == Curso{}) {
		p.CursoID = 0
	} else {
		p.CursoID = inscricao.Curso.ID
	}

	cpfRa, err := strconv.Atoi(inscricao.CpfRa)
	if err != nil {
		return err
	}

	p.CpfRa = cpfRa

	id, err := dbInsereParticipante(&p)
	if err != nil {
		return err
	}

	p.ID = int(id)

	// Inserir os minicursos
	for _, minicurso := range inscricao.Minicrusos {
		err = atualizaVagas(&minicurso, &p)
		if err != nil {
			return err
		}
	}
	return nil
}

func verificaInscricao(inscricao *ParticpanteRequest) (bool, error) {
	cpfRa, err := strconv.Atoi(inscricao.CpfRa)
	if err != nil {
		return false, err
	}

	participante, err := dbGetParticipante(cpfRa)
	if err != nil {
		return false, err
	}

	return participante != nil, nil
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

func PostNovaInscricao(c echo.Context) error {
	inscricao := new(ParticpanteRequest)

	if err := c.Bind(inscricao); err != nil {
		return err
	}

	existe, err := verificaInscricao(inscricao)
	if err != nil {
		fmt.Println("Could not verifica")
		return c.String(http.StatusBadRequest, "Erro ao verificar Usuario")
	}

	if existe {
		return c.String(http.StatusForbidden, "Iscricao ja existe")
	}

	err = novaInscricao(inscricao)
	if err != nil {
		fmt.Println("Failed to inscrever")
		return c.String(http.StatusBadRequest, "Erro ao realizar inscricao")
	}

	return nil
}

func login(c echo.Context) error {
	userJSON := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	if err := c.Bind(userJSON); err != nil {
		return err
	}

	if userJSON.Username == "admin" && userJSON.Password == "admin" {
		// Create token
		token := jwt.New(jwt.SigningMethodHS256)

		// Set claims
		claims := token.Claims.(jwt.MapClaims)
		claims["name"] = "admin"
		claims["admin"] = true

		// Set expiration date
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		// Gen the token
		t, err := token.SignedString([]byte("supersecret"))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, map[string]string{
			"token": t,
		})
	}

	return echo.ErrUnauthorized
}

func authBusinessRule(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	admin := claims["admin"].(bool)

	if admin {
		return c.String(http.StatusOK, "Hello admin: "+name+"!")
	}

	return echo.ErrUnauthorized
}

func authGetInscricao(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	admin := claims["admin"].(bool)

	if !admin {
		return echo.ErrUnauthorized
	}

	participantes, err := dbGetParticipantes()
	if err != nil {
		return err
	}

	if participantes != nil {
		return c.JSON(http.StatusOK, participantes)
	}

	fmt.Println("Participantes is null")
	return nil
}

func authCheckUser(c echo.Context) error {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	exp := claims["exp"].(float64)
	admin := claims["admin"].(bool)

	if int64(exp) >= time.Now().Unix() && !admin {
		// Token is invalid.
		return echo.ErrUnauthorized
	}
	return c.JSON(http.StatusOK, map[string]string{
		"msg": "Ok",
	})
}

func mudaPagamento(id int, foiPago bool) error {
	p, err := dbGetParticipanteId(id)
	if err != nil {
		return err
	}

	if p == nil {
		return fmt.Errorf("Nao encontrou participante id = %d", id)
	}

	p.Pago = foiPago

	// Atualiza pagamento.
	err = dbAtualizaParticipante(p)
	if err != nil {
		return err
	}

	return nil
}

func authFazPagamento(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	err := mudaPagamento(id, true)
	if err != nil {
		return err
	}
	return nil
}

func authDesfazPagamento(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	err := mudaPagamento(id, false)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// Echo instance
	e := echo.New()

	openDB, err := sql.Open("sqlite3", "./db/semana-zootecnia.db")
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

	e.GET("api/cursos", GetCursos)
	e.GET("api/minicursos", GetMiniCursos)
	e.POST("api/inscricao", PostNovaInscricao)
	e.POST("api/login", login)

	// Restricted api.
	r := e.Group("/api/protected")
	r.Use(middleware.JWT([]byte("supersecret")))
	r.GET("/business", authBusinessRule)
	r.GET("/check", authCheckUser)
	r.GET("/participantes", authGetInscricao)
	r.PUT("/fazpagamento/:id", authFazPagamento)
	r.PUT("/desfazpagamento/:id", authDesfazPagamento)

	// Start Server
	e.Logger.Fatal(e.Start(":1323"))
}
