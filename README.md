### Descrição

API para sistema de rifas em Golang, desenvolvido com IA

### Rotas

> Todas as rotas abaixo exigem `Authorization: Bearer <token>`. O "dono" da operação é
> sempre o usuário autenticado — não é mais informado na URL nem no corpo. A gestão de
> usuários (criar/atualizar/deletar conta) é feita diretamente no Clerk.

#### Rifas

| Método | Rota                    | Descrição                                       |
|--------|-------------------------|-------------------------------------------------|
| POST   | /raffles                | Criar rifa          |
| GET    | /raffles                | Listar rifas criadas pelo usuário    |
| GET    | /raffles/{raffle_id}    | Obter detalhe de uma rifa específica            |

#### Tickets

| Método | Rota       | Descrição                                                         |
|--------|------------|-------------------------------------------------------------------|
| POST   | /tickets   | Criar ticket      |
| GET    | /tickets   | Listar tickets (aceita query `?raffle_id=`) |

#### Webhooks

| Método | Rota              | Descrição                                              |
|--------|-------------------|--------------------------------------------------------|
| POST   | /webhooks/clerk   | Limpeza em cascata: no evento `user.deleted` do Clerk, remove as raffles/tickets do usuário |
