# 📊 Tabela de Testes da API ZMeow

| # | Funcionalidade | Endpoint | Método | Status | Detalhes |
|---|---|---|---|---|---|
| 1 | **Health Check** | `/health` | GET | ✅ **PASSOU** | Retorna status 200 com informações do sistema |
| 2 | **Listar Sessões** | `/sessions/list` | GET | ✅ **PASSOU** | Lista todas as sessões cadastradas |
| 3 | **Status da Sessão** | `/sessions/{id}/status` | GET | ✅ **PASSOU** | Retorna status de conexão da sessão |
| 4 | **Criar Sessão** | `/sessions/add` | POST | ✅ **PASSOU** | Cria nova sessão com sucesso |
| 5 | **QR Code** | `/sessions/{id}/qr` | GET | ✅ **PASSOU** | Retorna QR code para autenticação |
| 6 | **Envio de Texto** | `/messages/{id}/send/text` | POST | ✅ **PASSOU** | Mensagem enviada e entregue |
| 7 | **Envio de Localização** | `/messages/{id}/send/location` | POST | ✅ **PASSOU** | Localização enviada e entregue |
| 8 | **Envio de Contato** | `/messages/{id}/send/contact` | POST | ✅ **PASSOU** | Contato enviado e entregue |
| 9 | **Mensagem com Botões** | `/messages/{id}/send/buttons` | POST | ❌ **ERRO** | `server returned error 405` |
| 10 | **Mensagem com Lista** | `/messages/{id}/send/list` | POST | ❌ **ERRO** | `server returned error 405` |
| 11 | **Enquete** | `/messages/{id}/send/poll` | POST | ✅ **PASSOU** | Enviada com sucesso, recebendo votos |
| 12 | **Presença no Chat** | `/chat/{id}/presence` | POST | ✅ **PASSOU** | Status de digitação enviado |
| 13 | **Marcar como Lida** | `/chat/{id}/markread` | POST | ✅ **PASSOU** | Mensagens marcadas como lidas |
| 14 | **Listar Grupos** | `/groups/{id}/list` | GET | ✅ **PASSOU** | Lista grupos do usuário |
| 15 | **Info do Grupo** | `/groups/{id}/info` | GET | ✅ **PASSOU** | Retorna informações do grupo |
| 16 | **Obter Sessão** | `/sessions/{id}` | GET | ✅ **PASSOU** | Retorna detalhes completos da sessão |
| 17 | **Deletar Sessão** | `/sessions/{id}` | DELETE | ✅ **PASSOU** | Sessão removida com sucesso |
| 18 | **Conectar Sessão** | `/sessions/{id}/connect` | POST | ✅ **PASSOU** | Sessão conectada com sucesso |
| 19 | **Logout Sessão** | `/sessions/{id}/logout` | POST | ✅ **PASSOU** | Logout realizado com sucesso |
| 20 | **Pareamento por Telefone** | `/sessions/{id}/pairphone` | POST | ❌ **ERRO** | `Failed to pair phone` - erro interno |
| 21 | **Configurar Proxy** | `/sessions/{id}/proxy/set` | POST | ❌ **ERRO** | `Failed to set proxy` - erro interno |
| 22 | **Envio de Mídia** | `/messages/{id}/send/media` | POST | ✅ **PASSOU** | Mídia enviada com sucesso (corrigido campos JSON) |
| 23 | **Envio de Imagem** | `/messages/{id}/send/image` | POST | ✅ **PASSOU** | Imagem enviada com sucesso (corrigido campo `image`) |
| 24 | **Envio de Áudio** | `/messages/{id}/send/audio` | POST | ✅ **PASSOU** | Áudio enviado com sucesso (corrigido Base64) |
| 25 | **Envio de Vídeo** | `/messages/{id}/send/video` | POST | ✅ **PASSOU** | Vídeo enviado com sucesso (corrigido Base64) |
| 26 | **Envio de Documento** | `/messages/{id}/send/document` | POST | ✅ **PASSOU** | Documento enviado com sucesso (corrigido campos) |
| 27 | **Envio de Sticker** | `/messages/{id}/send/sticker` | POST | ❌ **ERRO** | `Failed to send sticker message` - erro interno |
| 28 | **Editar Mensagem** | `/messages/{id}/send/edit` | POST | ✅ **PASSOU** | Mensagem editada com sucesso |
| 29 | **Deletar Mensagem** | `/messages/{id}/delete` | POST | ✅ **PASSOU** | Mensagem deletada com sucesso |
| 30 | **Reagir Mensagem** | `/messages/{id}/react` | POST | ✅ **PASSOU** | Reação enviada e confirmada com sucesso |
| 31 | **Download de Imagem** | `/chat/{id}/downloadimage` | POST | ⏳ **PENDENTE** | Baixar imagem recebida |
| 32 | **Download de Vídeo** | `/chat/{id}/downloadvideo` | POST | ⏳ **PENDENTE** | Baixar vídeo recebido |
| 33 | **Download de Áudio** | `/chat/{id}/downloadaudio` | POST | ⏳ **PENDENTE** | Baixar áudio recebido |
| 34 | **Download de Documento** | `/chat/{id}/downloaddocument` | POST | ⏳ **PENDENTE** | Baixar documento recebido |
| 35 | **Criar Grupo** | `/groups/{id}/create` | POST | ✅ **PASSOU** | Grupo criado com sucesso, participantes adicionados |
| 36 | **Sair do Grupo** | `/groups/{id}/leave` | POST | ✅ **PASSOU** | Saiu do grupo com sucesso |
| 37 | **Atualizar Participantes** | `/groups/{id}/participants/update` | POST | ✅ **PASSOU** | Participantes adicionados com sucesso |
| 38 | **Definir Nome do Grupo** | `/groups/{id}/settings/name` | POST | ❌ **ERRO** | `Falha ao definir nome do grupo` - erro interno |
| 39 | **Definir Tópico do Grupo** | `/groups/{id}/settings/topic` | POST | ✅ **PASSOU** | Tópico definido com sucesso |
| 40 | **Definir Foto do Grupo** | `/groups/{id}/settings/photo` | POST | ❌ **ERRO** | `Falha ao definir foto do grupo` - erro interno |
| 41 | **Remover Foto do Grupo** | `/groups/{id}/settings/photo` | DELETE | ❌ **ERRO** | `Endpoint não encontrado` - não implementado |
| 42 | **Configurar Anúncios** | `/groups/{id}/settings/announce` | POST | ✅ **PASSOU** | Modo anúncio configurado com sucesso |
| 43 | **Bloquear Grupo** | `/groups/{id}/settings/locked` | POST | ✅ **PASSOU** | Modo bloqueado configurado com sucesso |
| 44 | **Mensagens Temporárias** | `/groups/{id}/settings/disappearing` | POST | ❌ **ERRO** | `Falha ao configurar timer` - erro interno |
| 45 | **Link de Convite** | `/groups/{id}/invite/link` | GET | ✅ **PASSOU** | Link obtido com sucesso |
| 46 | **Entrar via Link** | `/groups/{id}/invite/join` | POST | ✅ **PASSOU** | Entrou no grupo com sucesso |
| 47 | **Info do Convite** | `/groups/{id}/invite/info` | POST | ✅ **PASSOU** | Informações obtidas com sucesso |

## 📈 Estatísticas

| Métrica | Valor | Percentual |
|---|---|---|
| **Total de Funcionalidades** | 47 | 100% |
| **Testadas e Aprovadas** | 34 | 72% |
| **Testadas com Erro** | 13 | 28% |
| **Pendentes de Teste** | 0 | 0% |
| **Taxa de Sucesso (Testadas)** | 34/47 | **72%** |
