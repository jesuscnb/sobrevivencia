# Áudio inicial

Os arquivos WAV desta pasta são efeitos originais gerados por
`go run ./tools/audio_gen` na pasta `client`. O gerador usa apenas a biblioteca
padrão de Go e sementes fixas. Não há áudio de terceiros nem requisito de
atribuição.

Os sons de passos, coleta, construção, escavação, dano e vento estabelecem a
integração de áudio do jogo. Eles podem ser substituídos por gravações e mixagem
final sem alterar as regras das ações: `audio.go` é o catálogo e `config.go`
guarda os volumes padrão. `F10` silencia ou ativa o áudio durante a partida.
