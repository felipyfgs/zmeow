# ZMeow API - Documentação Completa

Esta documentação apresenta todas as rotas disponíveis na API ZMeow com exemplos de uso usando curl.

## Base URL
```
http://localhost:8080
```

## Índice
- [Health Check](#health-check)
- [Sessões](#sessões)
- [Mensagens](#mensagens)
- [Chat](#chat)
- [Grupos](#grupos)

---

## Health Check

### GET /health
Verifica a saúde da aplicação.

```bash
curl -X GET http://localhost:8080/health
```

**Resposta:**
```json
{
  "success": true,
  "message": "Service is healthy",
  "data": {
    "status": "ok",
    "service": "zmeow-api"
  }
}
```

---

## Sessões

### POST /sessions/add
Cria uma nova sessão WhatsApp.

```bash
curl -X POST http://localhost:8080/sessions/add \
  -H "Content-Type: application/json" \
  -d '{
    "name": "sessao-001",
    "webhook": "https://example.com/webhook",
    "proxyUrl": "socks5://proxy:1080"
  }'
```

### GET /sessions/list
Lista todas as sessões.

```bash
curl -X GET http://localhost:8080/sessions/list
```

### GET /sessions/{sessionID}
Obtém informações de uma sessão específica.

```bash
curl -X GET http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000
```

### DELETE /sessions/{sessionID}
Remove uma sessão.

```bash
curl -X DELETE http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000
```

### POST /sessions/{sessionID}/connect
Conecta uma sessão ao WhatsApp.

```bash
curl -X POST http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/connect
```

### POST /sessions/{sessionID}/logout
Desconecta uma sessão do WhatsApp.

```bash
curl -X POST http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/logout
```

### GET /sessions/{sessionID}/status
Obtém o status de uma sessão.

```bash
curl -X GET http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/status
```

### GET /sessions/{sessionID}/qr
Obtém o QR Code para autenticação.

```bash
curl -X GET http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/qr
```

### POST /sessions/{sessionID}/pairphone
Realiza pareamento via número de telefone.

```bash
curl -X POST http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/pairphone \
  -H "Content-Type: application/json" \
  -d '{
    "phoneNumber": "+5511999999999"
  }'
```

### POST /sessions/{sessionID}/proxy/set
Configura proxy para uma sessão.

```bash
curl -X POST http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/proxy/set \
  -H "Content-Type: application/json" \
  -d '{
    "proxyUrl": "socks5://proxy:1080"
  }'
```

---

## Mensagens

### POST /messages/{sessionID}/send/text
Envia mensagem de texto.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/text \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "message": "Olá! Esta é uma mensagem de teste.",
    "contextInfo": {
      "stanzaId": "message-id-to-reply",
      "mentionedJids": ["5511888888888@s.whatsapp.net"]
    }
  }'
```

### POST /messages/{sessionID}/send/media
Envia mídia unificada (imagem, áudio, vídeo, documento).

**Opção 1: JSON com Base64**
```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/media \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "mediaType": "image",
    "media": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD...",
    "caption": "Imagem enviada via API",
    "mimeType": "image/jpeg"
  }'
```

**Opção 2: Form-data para upload direto**
```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/media \
  -F "number=5511999999999" \
  -F "mediaType=image" \
  -F "caption=Imagem enviada via upload" \
  -F "media=@/path/to/image.jpg"
```

### POST /messages/{sessionID}/send/image
Envia imagem.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/image \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "image": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD...",
    "caption": "Legenda da imagem",
    "mimeType": "image/jpeg"
  }'
```

### POST /messages/{sessionID}/send/audio
Envia áudio.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/audio \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "audio": "data:audio/ogg;base64,T2dnUwACAAAAAAAAAADqnjMlAAAAAOyyzPIBHgF2b3JiaXMAAAAAAUAfAABAHwAAQB8AAEAfAACZAU9nZ1MAAAAAAAAAAAAA6p4zJQEAAAANJGeKCj3//////////5ADdm9yYmlzLQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAP...",
    "ptt": true,
    "caption": "Mensagem de voz"
  }'
```

### POST /messages/{sessionID}/send/video
Envia vídeo.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/video \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "video": "data:video/mp4;base64,AAAAIGZ0eXBpc29tAAACAGlzb21pc28yYXZjMW1wNDEAAAAIZnJlZQAACKBtZGF0AAAC...",
    "caption": "Vídeo enviado via API",
    "mimeType": "video/mp4"
  }'
```

### POST /messages/{sessionID}/send/document
Envia documento.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/document \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "document": "data:application/pdf;base64,JVBERi0xLjQKJcOkw7zDtsO8w6HDqMOgw6rDqcOlw6...",
    "fileName": "documento.pdf",
    "caption": "Documento importante",
    "mimeType": "application/pdf"
  }'
```

### POST /messages/{sessionID}/send/location
Envia localização.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/location \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "latitude": -23.5505,
    "longitude": -46.6333,
    "name": "São Paulo",
    "address": "São Paulo, SP, Brasil"
  }'
```

### POST /messages/{sessionID}/send/contact
Envia contato.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/contact \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "contactName": "João Silva",
    "contactJID": "5511888888888@s.whatsapp.net"
  }'
```

### POST /messages/{sessionID}/send/sticker
Envia sticker.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/sticker \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "sticker": "data:image/webp;base64,UklGRnoGAABXRUJQVlA4WAoAAAAQAAAAdwAABwAAQUxQSDIAAAARL0AmbZurmr57yyIiqE8oiG0bqdIgj...",
    "mimeType": "image/webp"
  }'
```

### POST /messages/{sessionID}/send/buttons
Envia mensagem com botões interativos.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/buttons \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "text": "Escolha uma opção:",
    "footer": "Powered by ZMeow",
    "buttons": [
      {
        "id": "btn1",
        "displayText": "Opção 1",
        "type": "RESPONSE"
      },
      {
        "id": "btn2",
        "displayText": "Opção 2",
        "type": "RESPONSE"
      }
    ]
  }'
```

### POST /messages/{sessionID}/send/list
Envia mensagem com lista interativa.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/list \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "text": "Selecione um item da lista:",
    "footer": "Powered by ZMeow",
    "title": "Menu Principal",
    "buttonText": "Ver Opções",
    "sections": [
      {
        "title": "Seção 1",
        "rows": [
          {
            "id": "item1",
            "title": "Item 1",
            "description": "Descrição do item 1"
          },
          {
            "id": "item2",
            "title": "Item 2",
            "description": "Descrição do item 2"
          }
        ]
      }
    ]
  }'
```

### POST /messages/{sessionID}/send/poll
Envia enquete.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/poll \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "name": "Qual sua cor favorita?",
    "options": ["Azul", "Verde", "Vermelho", "Amarelo"],
    "selectableOptionsCount": 1
  }'
```

### POST /messages/{sessionID}/send/edit
Edita uma mensagem enviada.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/edit \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "id": "message-id-to-edit",
    "newText": "Texto editado da mensagem"
  }'
```

### POST /messages/{sessionID}/delete
Deleta uma mensagem.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/delete \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "id": "message-id-to-delete",
    "forMe": false
  }'
```

### POST /messages/{sessionID}/react
Reage a uma mensagem.

```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/react \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "id": "message-id-to-react",
    "reaction": "👍"
  }'
```

---

## Chat

### POST /chat/{sessionID}/presence
Define presença no chat (digitando, gravando, etc.).

```bash
curl -X POST http://localhost:8080/chat/550e8400-e29b-41d4-a716-446655440000/presence \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "5511999999999",
    "state": "composing"
  }'
```

### POST /chat/{sessionID}/markread
Marca mensagens como lidas.

```bash
curl -X POST http://localhost:8080/chat/550e8400-e29b-41d4-a716-446655440000/markread \
  -H "Content-Type: application/json" \
  -d '{
    "message_ids": ["msg1", "msg2", "msg3"],
    "chat": "5511999999999@s.whatsapp.net"
  }'
```

### POST /chat/{sessionID}/downloadimage
Faz download de uma imagem.

```bash
curl -X POST http://localhost:8080/chat/550e8400-e29b-41d4-a716-446655440000/downloadimage \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": "image-message-id",
    "phone": "5511999999999"
  }'
```

### POST /chat/{sessionID}/downloadvideo
Faz download de um vídeo.

```bash
curl -X POST http://localhost:8080/chat/550e8400-e29b-41d4-a716-446655440000/downloadvideo \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": "video-message-id",
    "phone": "5511999999999"
  }'
```

### POST /chat/{sessionID}/downloadaudio
Faz download de um áudio.

```bash
curl -X POST http://localhost:8080/chat/550e8400-e29b-41d4-a716-446655440000/downloadaudio \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": "audio-message-id",
    "phone": "5511999999999"
  }'
```

### POST /chat/{sessionID}/downloaddocument
Faz download de um documento.

```bash
curl -X POST http://localhost:8080/chat/550e8400-e29b-41d4-a716-446655440000/downloaddocument \
  -H "Content-Type: application/json" \
  -d '{
    "message_id": "document-message-id",
    "phone": "5511999999999"
  }'
```

---

## Grupos

### POST /groups/{sessionID}/create
Cria um novo grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/create \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Meu Grupo",
    "participants": ["5511999999999", "5511888888888"]
  }'
```

### GET /groups/{sessionID}/list
Lista todos os grupos da sessão.

```bash
curl -X GET http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/list
```

### GET /groups/{sessionID}/info
Obtém informações de um grupo específico.

```bash
curl -X GET "http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/info?group_jid=120363025246125486@g.us"
```

### POST /groups/{sessionID}/leave
Sai de um grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/leave \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us"
  }'
```

### POST /groups/{sessionID}/participants/update
Atualiza participantes do grupo (adicionar, remover, promover, rebaixar).

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/participants/update \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "phones": ["5511999999999"],
    "action": "add"
  }'
```

### POST /groups/{sessionID}/settings/name
Define o nome do grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/name \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "name": "Novo Nome do Grupo"
  }'
```

### POST /groups/{sessionID}/settings/topic
Define o tópico/descrição do grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/topic \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "topic": "Descrição do grupo atualizada"
  }'
```

### POST /groups/{sessionID}/settings/photo
Define a foto do grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/photo \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "image": "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD..."
  }'
```

### DELETE /groups/{sessionID}/settings/photo
Remove a foto do grupo.

```bash
curl -X DELETE http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/photo \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us"
  }'
```

### POST /groups/{sessionID}/settings/announce
Configura se apenas admins podem enviar mensagens.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/announce \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "announce": true
  }'
```

### POST /groups/{sessionID}/settings/locked
Configura se apenas admins podem editar informações do grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/locked \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "locked": true
  }'
```

### POST /groups/{sessionID}/settings/disappearing
Configura timer de mensagens que desaparecem.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/settings/disappearing \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "group_jid": "120363025246125486@g.us",
    "duration": "7d"
  }'
```

### GET /groups/{sessionID}/invite/link
Obtém o link de convite do grupo.

```bash
curl -X GET "http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/invite/link?group_jid=120363025246125486@g.us&reset=false"
```

### POST /groups/{sessionID}/invite/join
Entra em um grupo usando link de convite.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/invite/join \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "invite-code-here"
  }'
```

### POST /groups/{sessionID}/invite/info
Obtém informações de um convite de grupo.

```bash
curl -X POST http://localhost:8080/groups/550e8400-e29b-41d4-a716-446655440000/invite/info \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "invite-code-here"
  }'
```

---

## Códigos de Resposta HTTP

| Código | Descrição |
|--------|-----------|
| 200 | Sucesso |
| 201 | Criado com sucesso |
| 400 | Requisição inválida |
| 404 | Recurso não encontrado |
| 500 | Erro interno do servidor |

## Formato de Resposta Padrão

### Resposta de Sucesso
```json
{
  "success": true,
  "message": "Operação realizada com sucesso",
  "data": {
    // dados específicos da operação
  }
}
```

### Resposta de Erro
```json
{
  "success": false,
  "message": "Descrição do erro",
  "error": {
    "code": "ERROR_CODE",
    "details": "Detalhes específicos do erro"
  }
}
```

## Notas Importantes

### Formatos de Telefone
- Use números no formato internacional sem o sinal de +
- Exemplo: `5511999999999` para +55 11 99999-9999

### Formatos de Mídia
- **Base64**: `data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD...`
- **URL**: `https://example.com/image.jpg`
- **Upload direto**: Use form-data com campo `media`

### IDs de Sessão
- Todos os IDs de sessão devem ser UUIDs válidos
- Exemplo: `550e8400-e29b-41d4-a716-446655440000`

### Estados de Presença
- `composing`: Digitando
- `recording`: Gravando áudio
- `paused`: Pausado

### Ações de Participantes
- `add`: Adicionar participante
- `remove`: Remover participante
- `promote`: Promover a admin
- `demote`: Rebaixar de admin

### Durações de Mensagens que Desaparecem
- `off`: Desabilitado
- `24h`: 24 horas
- `7d`: 7 dias
- `90d`: 90 dias

## Exemplos de Uso Completo

### Fluxo Básico: Criar Sessão e Enviar Mensagem

1. **Criar sessão:**
```bash
curl -X POST http://localhost:8080/sessions/add \
  -H "Content-Type: application/json" \
  -d '{"name": "minha-sessao"}'
```

2. **Conectar sessão:**
```bash
curl -X POST http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/connect
```

3. **Obter QR Code:**
```bash
curl -X GET http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/qr
```

4. **Verificar status:**
```bash
curl -X GET http://localhost:8080/sessions/550e8400-e29b-41d4-a716-446655440000/status
```

5. **Enviar mensagem:**
```bash
curl -X POST http://localhost:8080/messages/550e8400-e29b-41d4-a716-446655440000/send/text \
  -H "Content-Type: application/json" \
  -d '{
    "number": "5511999999999",
    "message": "Olá! Mensagem enviada via API ZMeow"
  }'
```

### Testando a API

Para testar todas as rotas rapidamente, você pode usar o seguinte script bash:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"
SESSION_ID="550e8400-e29b-41d4-a716-446655440000"

echo "=== Testando Health Check ==="
curl -X GET $BASE_URL/health

echo -e "\n\n=== Criando Sessão ==="
curl -X POST $BASE_URL/sessions/add \
  -H "Content-Type: application/json" \
  -d '{"name": "teste-api"}'

echo -e "\n\n=== Listando Sessões ==="
curl -X GET $BASE_URL/sessions/list

echo -e "\n\n=== Conectando Sessão ==="
curl -X POST $BASE_URL/sessions/$SESSION_ID/connect

echo -e "\n\n=== Verificando Status ==="
curl -X GET $BASE_URL/sessions/$SESSION_ID/status
```

---

## Suporte

Para mais informações sobre a API ZMeow, consulte:
- Código fonte: `/root/zmeow`
- Logs da aplicação para debugging
- Documentação do WhatsApp Business API para referências adicionais
