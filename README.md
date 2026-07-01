### Descrição

API para sistema de rifas em Golang, desenvolvido com IA

### Rotas

#### Usuários

| Método | Rota          | Descrição          |
|--------|---------------|--------------------|
| POST   | /users        | Criar usuário      |
| PUT    | /users/{id}   | Atualizar usuário  |
| DELETE | /users/{id}   | Deletar usuário    |

#### Rifas

| Método | Rota                              | Descrição                            |
|--------|-----------------------------------|--------------------------------------|
| POST   | /raffles                          | Criar rifa                           |
| GET    | /users/{id}/raffles               | Listar rifas de um usuário           |
| GET    | /users/{id}/raffles/{raffle_id}   | Obter detalhe de uma rifa específica |

#### Tickets

| Método | Rota                  | Descrição                                                 |
|--------|-----------------------|-----------------------------------------------------------|
| POST   | /tickets              | Criar ticket                                              |
| GET    | /users/{id}/tickets   | Listar tickets de um usuário (aceita query `?raffle_id=`) |