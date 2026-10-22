# Texturas do terreno

`grass.png` e `soil.png` são texturas originais geradas pelo próprio projeto com
`go run ./tools/terrain_textures.go` a partir da pasta `client`. O gerador usa
ruído periódico e semente fixa: as bordas de cada imagem se repetem e os
arquivos podem ser recriados sem baixar recursos externos.

As duas imagens têm 128 × 128 pixels e são compartilhadas por todos os chunks.
São a primeira direção visual estilizada para o chão; cores e detalhes podem
ser refinados depois sem mudar o sistema de projeção ou o formato do save.
