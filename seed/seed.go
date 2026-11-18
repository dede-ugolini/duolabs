package seed

import (
	"errors"
	"log"

	"duolabs/model"

	"gorm.io/gorm"
)

// legacyQuestions are the one-question sample lessons that shipped before the
// sample content was rewritten. Upgrade removes them. A current
// sample question must never equal one of these, or it would be deleted on
// every start (main_test.go enforces this).
var legacyQuestions = []string{
	"Como declarar uma variável constante em Go?",
	"Qual função do Pandas é usada para ler um arquivo CSV?",
	"Qual operador utilitário torna todas as propriedades de um tipo opcionais?",
}

// newLesson builds a lesson from the choice labels in display order;
// correct is the index of the right one.
func newLesson(question string, correct int, labels ...string) model.Lesson {
	choices := make([]model.Choice, len(labels))
	for i, label := range labels {
		choices[i] = model.Choice{Label: label, Correct: i == correct}
	}
	return model.Lesson{Question: question, Choices: choices}
}

func newTask(title string, lessons ...model.Lesson) model.Task {
	return model.Task{Title: title, Lessons: lessons}
}

func newUnit(title string, tasks ...model.Task) model.Unit {
	return model.Unit{Title: title, Tasks: tasks}
}

func sampleCourses() []model.Course {
	return []model.Course{
		{
			Title:       "Go Fundamentals",
			Description: "Aprenda a linguagem Go do zero ao avançado com concorrência e microsserviços.",
			Category:    "Programação",
			Difficulty:  model.Beginner,
			Xp:          500,
			Units: []model.Unit{
				newUnit("Módulo 1: Fundamentos da linguagem",
					newTask("Variáveis e constantes",
						newLesson("Qual declaração cria uma constante em Go?", 1,
							"var limite = 10",
							"const limite = 10",
							"let limite = 10",
							"final int limite = 10"),
						newLesson("O que o operador := faz em Go?", 2,
							"Compara dois valores por igualdade",
							"Declara uma constante",
							"Declara e inicializa uma variável, inferindo o tipo (só dentro de funções)",
							"Cria um ponteiro para a variável"),
						newLesson("Qual é o valor de s logo após var s string, sem inicializar?", 0,
							`String vazia ("")`,
							"nil",
							`"null"`,
							"Erro de compilação: variável não inicializada"),
					),
					newTask("Funções",
						newLesson("Qual assinatura declara uma função que retorna um int e um error?", 3,
							"func dividir(a, b int) int, error",
							"func dividir(a, b int) -> (int, error)",
							"function dividir(a, b int): (int, error)",
							"func dividir(a, b int) (int, error)"),
						newLesson("O que a instrução defer faz?", 1,
							"Executa a chamada em outra goroutine",
							"Adia a chamada para quando a função atual retornar",
							"Ignora qualquer erro da chamada",
							"Cancela a execução da função atual"),
					),
				),
				newUnit("Módulo 2: Coleções e tipos",
					newTask("Slices e maps",
						newLesson("Qual expressão cria um slice de int com tamanho 3 e capacidade 5?", 2,
							"make([]int, 5, 3)",
							"new([]int, 3, 5)",
							"make([]int, 3, 5)",
							"[]int{3, 5}"),
						newLesson(`Como saber se a chave "x" existe no map m?`, 0,
							`v, ok := m["x"] e testar ok`,
							`m.has("x")`,
							`m.contains("x")`,
							`if m["x"] != nil`),
					),
					newTask("Structs e interfaces",
						newLesson("Quando um tipo satisfaz uma interface em Go?", 3,
							"Quando usa a palavra-chave implements",
							"Quando herda de um tipo base que declara a interface",
							"Quando a interface é registrada em uma função init()",
							"Quando implementa todos os métodos da interface, sem declarar isso explicitamente"),
					),
				),
				newUnit("Módulo 3: Concorrência",
					newTask("Goroutines e canais",
						newLesson("Como executar a função f() em uma nova goroutine?", 2,
							"async f()",
							"thread f()",
							"go f()",
							"spawn f()"),
						newLesson("O que acontece ao executar ch <- 1 em um canal sem buffer quando nenhuma goroutine está recebendo?", 1,
							"O valor é descartado",
							"A goroutine que envia bloqueia até que alguém receba",
							"O programa entra em panic imediatamente",
							"O valor fica em uma fila interna até alguém ler"),
						newLesson("Para que serve sync.WaitGroup?", 0,
							"Esperar que um conjunto de goroutines termine",
							"Proteger uma variável com exclusão mútua",
							"Limitar o tempo de execução de uma goroutine",
							"Criar um pool de threads do sistema operacional"),
					),
				),
			},
		},
		{
			Title:       "Python para Análise de Dados",
			Description: "Domine Pandas, NumPy e visualização de dados com Python.",
			Category:    "Ciência de Dados",
			Difficulty:  model.Intermediate,
			Xp:          750,
			Units: []model.Unit{
				newUnit("Módulo 1: Pandas básico",
					newTask("Lendo e inspecionando dados",
						newLesson("Qual função do pandas carrega um arquivo CSV em um DataFrame?", 3,
							`pd.open_csv("dados.csv")`,
							`pd.load_csv("dados.csv")`,
							`pd.import_csv("dados.csv")`,
							`pd.read_csv("dados.csv")`),
						newLesson("Qual método exibe as 5 primeiras linhas de um DataFrame df?", 1,
							"df.top()",
							"df.head()",
							"df.peek()",
							"df.limit(5)"),
						newLesson("Qual método resume média, desvio padrão, mínimo e máximo das colunas numéricas?", 2,
							"df.summary()",
							"df.stats()",
							"df.describe()",
							"df.info()"),
					),
					newTask("Selecionando e filtrando",
						newLesson("Como selecionar apenas as linhas em que a coluna idade é maior que 30?", 0,
							`df[df["idade"] > 30]`,
							"df.where(idade > 30)",
							`df.filter("idade > 30")`,
							`df.select(df["idade"] > 30)`),
						newLesson("Qual a diferença entre df.loc[...] e df.iloc[...]?", 1,
							"loc só aceita colunas e iloc só aceita linhas",
							"loc seleciona por rótulo (nomes do índice e das colunas) e iloc por posição inteira",
							"loc seleciona por posição inteira e iloc por rótulo",
							"Não há diferença: são apelidos do mesmo método"),
					),
				),
				newUnit("Módulo 2: Limpeza e agregação",
					newTask("Valores ausentes",
						newLesson("Qual método devolve um DataFrame de booleanos indicando quais valores estão ausentes (NaN)?", 2,
							"df.isempty()",
							"df.missing()",
							"df.isna()",
							"df.nulls()"),
						newLesson("Qual chamada preenche com 0 os valores ausentes da coluna nota?", 0,
							`df["nota"].fillna(0)`,
							`df["nota"].replace_na(0)`,
							`df["nota"].fill(0)`,
							`df["nota"].dropna(0)`),
					),
					newTask("Agrupamento e junção",
						newLesson("Como calcular a média de salario para cada departamento?", 3,
							`df.mean("departamento", "salario")`,
							`df.groupby("salario")["departamento"].mean()`,
							`df["salario"].groupby().mean()`,
							`df.groupby("departamento")["salario"].mean()`),
						newLesson("Qual função junta dois DataFrames a partir de uma coluna em comum, como um JOIN de SQL?", 1,
							"pd.concat([a, b])",
							`pd.merge(a, b, on="id")`,
							`pd.append(a, b, on="id")`,
							`pd.join_on(a, b, "id")`),
					),
				),
			},
		},
		{
			Title:       "TypeScript Avançado",
			Description: "Tipos avançados, generics, utilitários e metaprogramação em TypeScript.",
			Category:    "Desenvolvimento Web",
			Difficulty:  model.Advanced,
			Xp:          1000,
			Units: []model.Unit{
				newUnit("Módulo 1: Tipos utilitários e genéricos",
					newTask("Tipos utilitários",
						newLesson("Qual tipo utilitário torna todas as propriedades de T opcionais?", 2,
							"Optional<T>",
							"Maybe<T>",
							"Partial<T>",
							"Nullable<T>"),
						newLesson("Qual tipo utilitário constrói um tipo contendo apenas as propriedades K de T?", 0,
							"Pick<T, K>",
							"Omit<T, K>",
							"Extract<T, K>",
							"Exclude<T, K>"),
						newLesson("O que o tipo Record<string, number> representa?", 3,
							"Uma tupla com uma string e um número",
							"Um array de pares [string, number]",
							"Uma classe com dois campos",
							"Um objeto cujas chaves são strings e cujos valores são números"),
					),
					newTask("Genéricos",
						newLesson("Qual declaração restringe T a tipos que possuem a propriedade length?", 1,
							"function f<T implements { length: number }>(x: T)",
							"function f<T extends { length: number }>(x: T)",
							"function f<T: { length: number }>(x: T)",
							"function f<T = { length: number }>(x: T)"),
						newLesson("Para que serve a palavra-chave infer em tipos condicionais?", 2,
							"Converter um valor para outro tipo em tempo de execução",
							"Desativar a checagem de tipos de uma expressão",
							"Capturar um tipo dentro de extends para usá-lo no resultado, como em T extends Promise<infer U> ? U : T",
							"Declarar uma variável cujo tipo vem do valor inicial"),
					),
				),
				newUnit("Módulo 2: Tipos mapeados e narrowing",
					newTask("Tipos mapeados",
						newLesson("O que faz o tipo { readonly [K in keyof T]: T[K] }?", 0,
							"Cria um tipo com as mesmas propriedades de T, todas somente leitura",
							"Cria um tipo apenas com os nomes das chaves de T",
							"Torna todas as propriedades de T opcionais",
							"Remove todas as propriedades de T"),
						newLesson("Qual é o resultado de keyof { id: number; nome: string }?", 3,
							"number | string",
							`["id", "nome"]`,
							"string",
							`"id" | "nome"`),
					),
					newTask("Narrowing",
						newLesson("Qual o papel de x is string como tipo de retorno de uma função (type predicate)?", 1,
							"Converter x em string em tempo de execução",
							"Dizer ao compilador que, se a função retornar true, x pode ser tratado como string naquele escopo",
							"Garantir que a função nunca lance exceções",
							"Fazer a função retornar sempre uma string"),
						newLesson("Em uma união discriminada, o que permite ao TypeScript restringir o tipo dentro de um switch?", 2,
							"O operador typeof aplicado ao objeto inteiro",
							"Anotações any em cada case",
							`Uma propriedade literal comum a todos os membros da união, como kind: "circle"`,
							"A ordem em que as propriedades foram declaradas"),
					),
				),
			},
		},
	}
}

// Run seeds the sample courses into an empty database.
func Run(db *gorm.DB) error {
	var count int64
	if err := db.Model(&model.Course{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, c := range sampleCourses() {
		if err := db.Create(&c).Error; err != nil {
			return err
		}
	}
	log.Println("Database seeded with sample courses successfully!")
	return nil
}

// Upgrade brings databases seeded with the old one-question
// content up to date, idempotently:
//  1. deletes the legacy sample lessons (with their choices and answers) from
//     every course;
//  2. refills a sample course (matched by title) that is left with no lessons.
//
// A sample course the admin deleted is not recreated, and one that still has
// lessons is left alone.
func Upgrade(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var legacyIDs []uint
		if err := tx.Model(&model.Lesson{}).Where("question IN ?", legacyQuestions).Pluck("id", &legacyIDs).Error; err != nil {
			return err
		}
		if len(legacyIDs) > 0 {
			for _, table := range []any{&model.Choice{}, &model.Answer{}} {
				if err := tx.Unscoped().Where("lesson_id IN ?", legacyIDs).Delete(table).Error; err != nil {
					return err
				}
			}
			if err := tx.Unscoped().Where("id IN ?", legacyIDs).Delete(&model.Lesson{}).Error; err != nil {
				return err
			}
		}

		for _, sample := range sampleCourses() {
			var course model.Course
			if err := tx.Where("title = ?", sample.Title).First(&course).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			var lessons int64
			if err := tx.Table("lessons").
				Joins("JOIN tasks ON tasks.id = lessons.task_id AND tasks.deleted_at IS NULL").
				Joins("JOIN units ON units.id = tasks.unit_id AND units.deleted_at IS NULL").
				Where("units.course_id = ? AND lessons.deleted_at IS NULL", course.ID).
				Count(&lessons).Error; err != nil {
				return err
			}
			if lessons > 0 {
				continue
			}
			unitIDs := tx.Unscoped().Model(&model.Unit{}).Select("id").Where("course_id = ?", course.ID)
			if err := tx.Unscoped().Where("unit_id IN (?)", unitIDs).Delete(&model.Task{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("course_id = ?", course.ID).Delete(&model.Unit{}).Error; err != nil {
				return err
			}
			for _, unit := range sample.Units {
				unit.CourseID = course.ID
				if err := tx.Create(&unit).Error; err != nil {
					return err
				}
			}
			log.Printf("Sample course %q updated with new content", course.Title)
		}
		return nil
	})
}
