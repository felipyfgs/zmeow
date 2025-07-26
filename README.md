# ZMeow API

API REST em Go para gerenciamento de múltiplas sessões WhatsApp usando a biblioteca whatsmeow.

## 🚀 Características

- **Clean Architecture**: Implementação seguindo os princípios da Clean Architecture
- **Múltiplas Sessões**: Gerenciamento de várias sessões WhatsApp simultaneamente
- **API REST**: Interface HTTP completa para todas as operações
- **QR Code e Phone Pairing**: Suporte a autenticação por QR Code e pareamento por telefone
- **Rate Limiting**: Controle de taxa de requisições
- **Logging Estruturado**: Sistema de logs usando zerolog
- **Database Migrations**: Sistema automático de migrações

## 📋 Pré-requisitos

- Go 1.23.4 ou superior
- PostgreSQL 12 ou superior
- Git

## 🛠️ Instalação

1. **Clone o repositório:**
```bash
git clone <repository-url>
cd zmeow
```

2. **Configure o banco de dados PostgreSQL:**
```sql
CREATE DATABASE zmeow_db;
CREATE USER zmeow WITH PASSWORD 'zmeow_password';
GRANT ALL PRIVILEGES ON DATABASE zmeow_db TO zmeow;
```

3. **Configure as variáveis de ambiente:**
```bash
cp .env.example .env
# Edite o arquivo .env com suas configurações
```

4. **Instale as dependências:**
```bash
go mod download
```

5. **Execute as migrações:**
```bash
go run cmd/zmeow/main.go
```

## ⚡ Executando

```bash
go run cmd/zmeow/main.go
```

A API estará disponível em `http://localhost:8080`

## 📚 Endpoints da API

### Gerenciamento de Sessões

#### 1. Criar Sessão
```http
POST /sessions/add
Content-Type: application/json

{
    "name": "sessao-001",
    "webhook": "https://example.com/webhook",
    "proxyUrl": "socks5://proxy:1080"
}
```

#### 2. Listar Sessões
```http
GET /sessions/list
```

#### 3. Obter Sessão
```http
GET /sessions/{sessionID}
```

#### 4. Deletar Sessão
```http
DELETE /sessions/{sessionID}
```

### Operações de Conectividade

#### 5. Conectar Sessão
```http
POST /sessions/{sessionID}/connect
```

#### 6. Desconectar Sessão
```http
POST /sessions/{sessionID}/logout
```

#### 7. Status da Sessão
```http
GET /sessions/{sessionID}/status
```

### Autenticação

#### 8. QR Code
```http
GET /sessions/{sessionID}/qr
```

#### 9. Pareamento por Telefone
```http
POST /sessions/{sessionID}/pairphone
Content-Type: application/json

{
    "phoneNumber": "+5511999999999"
}
```

### Configuração

#### 10. Configurar Proxy
```http
POST /sessions/{sessionID}/proxy/set
Content-Type: application/json

{
    "proxyUrl": "socks5://user:pass@proxy.com:1080"
}
```

### Health Check

#### 11. Health Check
```http
GET /health
```

## 📦 Estrutura do Projeto

```
zmeow/
├── cmd/
│   └── zmeow/
│       └── main.go                 # Entry point
├── internal/
│   ├── domain/                     # Entidades e regras de negócio
│   │   ├── session/
│   │   └── whatsapp/
│   ├── usecases/                   # Casos de uso
│   │   └── session/
│   ├── infra/                      # Implementações externas
│   │   ├── database/
│   │   └── whatsapp/
│   ├── http/                       # HTTP handlers e middleware
│   │   ├── handlers/
│   │   ├── middleware/
│   │   ├── router/
│   │   └── responses/
│   └── app/                        # Configuração da aplicação
│       ├── config/
│       ├── dependency/
│       └── server/
├── pkg/                            # Bibliotecas reutilizáveis
│   └── logger/
└── docs/                           # Documentação
```

## 🔧 Configuração

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `APP_ENV` | Ambiente da aplicação | `development` |
| `APP_HOST` | Host do servidor | `localhost` |
| `APP_PORT` | Porta do servidor | `8080` |
| `DB_HOST` | Host do PostgreSQL | `localhost` |
| `DB_PORT` | Porta do PostgreSQL | `5432` |
| `DB_USER` | Usuário do banco | `zmeow` |
| `DB_PASSWORD` | Senha do banco | - |
| `DB_NAME` | Nome do banco | `zmeow_db` |
| `DB_SSLMODE` | Modo SSL do banco | `disable` |
| `LOG_LEVEL` | Nível de log | `info` |
| `LOG_FORMAT` | Formato do log | `console` |

## 🚀 Deploy

### Docker

1. **Build da imagem:**
```bash
docker build -t zmeow:latest .
```

2. **Executar com docker-compose:**
```bash
docker-compose up -d
```

## 📝 Exemplos de Uso

### Criar e Conectar uma Sessão

1. **Criar sessão:**
```bash
curl -X POST http://localhost:8080/sessions/add \
  -H "Content-Type: application/json" \
  -d '{"name": "minha-sessao"}'
```

2. **Conectar sessão:**
```bash
curl -X POST http://localhost:8080/sessions/{sessionID}/connect
```

3. **Obter QR Code:**
```bash
curl http://localhost:8080/sessions/{sessionID}/qr
```

## 🔍 Logs

Os logs são estruturados e incluem:
- Timestamp
- Nível (DEBUG, INFO, WARN, ERROR)
- Componente
- Contexto (sessionID, operation, etc.)
- Mensagem

## 🤝 Contribuição

1. Fork o projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Push para a branch
5. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para mais detalhes.

## 🐛 Issues

Para reportar bugs ou solicitar features, use o sistema de issues do GitHub.

## 📞 Suporte

Para suporte, entre em contato através dos issues do GitHub.