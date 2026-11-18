# Duolabs

Trabalho da matéria de desenvolvimento Web da faculdade: plataforma de cursos estilo Duolingo (cursos > módulos > tarefas > lições com múltipla escolha, matrícula e progresso).

## Como rodar

Docker (backend + frontend juntos, em http://localhost:8081):

```
docker build -t duolabs .
docker run -p 8081:8081 -v duolabs-data:/data -e JWT_SECRET=troque-isto -e ADMIN_PASSWORD=troque-isto duolabs:latest
```

Desenvolvimento (dois terminais):

```
go run main.go                       # API em :8081
cd frontend && npm install && npm run dev   # front em :5173
```

Testes: `go vet ./... && go test ./...`

## Variáveis de ambiente

| Variável | Padrão | Para quê |
| --- | --- | --- |
| `PORT` | `8081` | porta do servidor |
| `DB_PATH` | `duolabs.db` | arquivo SQLite |
| `STATIC_DIR` | `static` | pasta do front compilado |
| `JWT_SECRET` | chave de desenvolvimento | chave que assina os tokens; **defina em produção**, senão qualquer um forja um token de admin |
| `ADMIN_NAME` / `ADMIN_PASSWORD` | `admin` / `admin123` | conta de administrador, criada a cada inicialização pelo `seed` |

## Estrutura

```
main.go          liga tudo: banco, seed, rotas
model/           structs do banco e do JSON (User, Course, Unit, Task, Lesson, Choice, ...)
store/           acesso ao banco (uma interface + implementação SQLite por entidade)
handler/         handlers HTTP, middlewares de login/admin e helpers (writeJSON, pathID, ...)
seed/            dados iniciais: cursos de exemplo e a conta de administrador
frontend/src/
  api.js         todas as chamadas à API (os componentes nunca usam fetch direto)
  auth.js        token no localStorage e leitura do JWT
  courseForm.js  validação e payload do editor de cursos (sem React)
  components/    telas: AuthForm, CourseList, EnrollmentList, CoursePlayer, CourseEditor...
swagger.yaml     documentação da API
```

## Permissões

- Qualquer pessoa vê os cursos; só quem está logado se matricula e responde lições.
- Criar, editar e apagar cursos exige o papel `admin` (checado no servidor por `AdminMiddleware`; o front só esconde os botões).
- O gabarito (`correct`) só é enviado ao admin.
- Só o dono da matrícula (ou um admin) pode cancelá-la.
