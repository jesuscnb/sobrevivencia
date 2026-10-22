# Design: mundo expansível e terreno texturizado

**Estado:** desenho revisado e aprovado pelo autor em 21/10/2026; pronto para planejamento por entregas.

## Objetivo e limites

O mundo deve crescer conforme o jogador se desloca, carregando e liberando regiões sem interromper a travessia. O terreno terá aparência low poly estilizada, com textura e iluminação coerentes com os assets existentes. O mesmo sistema deve permitir estradas, assentamentos e minimapa contínuos entre regiões. O autor continuará escrevendo o código com orientação didática e checkpoints visíveis.

Nesta fase não serão implementados banco de dados, agente de IA, cidades completas, multiplayer, níveis de detalhe nem uma navegação nova para inimigos. Eles são requisitos de evolução para as estruturas de dados, não dependências da primeira travessia entre regiões. “Infinito” significa gerar coordenadas sob demanda e manter somente uma vizinhança ativa; não significa memória ou coordenadas matematicamente ilimitadas.

## Situação atual e reaproveitamento

- `Mundo` guarda um campo `[17][9][17]float32` e um `rl.Model` para o terreno inteiro. O volume cobre inicialmente X/Z de −8 a 8 e Y de −4 a 4.
- `terrain_model.go` extrai uma malha com Polyform, escreve um OBJ temporário e o carrega com raylib. Escavação reconstrói a malha inteira.
- Raycasts de jogador, mira, recursos, construção e inimigos consultam esse único modelo. A navegação A* e o limite do jogador também usam os limites fixos da região inicial.
- `RegiaoGeracao`, `IDObjeto` e a semente por região já preparam IDs reproduzíveis para objetos de superfície. `IndiceOcupacao` usa posições de mundo e pode continuar como índice das instâncias carregadas.
- O save v4 armazena alterações do campo inicial e o estado de recursos, construções e inimigos. Esse formato precisa de leitura compatível durante a migração.

## Coordenadas e regiões de volume

Reaproveitar inicialmente células de 16 × 8 × 16, com 17 × 9 × 17 amostras de densidade. Uma coordenada inteira de região `(X, Y, Z)` identifica o volume; só a camada da superfície precisa estar ativa no primeiro checkpoint. Incluir `Y` na identidade preserva a possibilidade de volumes acima e abaixo para cavernas e escavação futura. As regiões de distribuição de árvores, rochas, estradas e assentamentos continuam organizadas em X/Z.

A amostra de densidade é endereçada por **coordenada global inteira**. A última amostra de uma região e a primeira da vizinha apontam para o mesmo ponto global. Ruído, edições e cálculo de normais consultam essa fonte única para evitar fendas geométricas e diferenças de iluminação. A conversão de posição para região deve funcionar com coordenadas negativas, usando divisão por piso em vez de truncamento em direção a zero.

Cada região ativa possui campo, malha/modelo, estado de carregamento e instâncias locais. `Mundo` mantém um mapa de regiões ativas e oferece consultas de terreno aos sistemas existentes. O restante do jogo não deve escolher diretamente `modeloTerreno` ou `campo` de uma região.

## Geração, carregamento e escavação

O estado base é função da semente, versão do gerador e coordenadas globais. Ao ativar uma região, aplicar as alterações persistidas antes de extrair a malha e posicionar seus objetos. A geração não pode depender da ordem em que regiões são carregadas. O jogador deve ter a região de destino pronta antes de atravessar a borda; a ausência de malha ou colisão não pode ser tratada como chão livre.

O carregador prepara regiões próximas, estabelece um limite de trabalho por quadro e libera modelos e instâncias distantes. Geração de densidades e preparação de geometria podem ocorrer fora do desenho; criação e liberação de recursos gráficos seguem as exigências do contexto da raylib. Uma margem diferente para carregar e descarregar evita alternância repetida na borda.

Um corte altera amostras globais, marca todas as regiões que compartilham os pontos modificados e reconstrói somente as malhas afetadas. Raycasts e colisão devem consultar a região correta após a troca de malha. Antes de usar muitas regiões, validar um caminho de criação de malha sem OBJ temporário e, especificamente, a propriedade e liberação da memória de vértices na versão instalada de raylib-go. O OBJ temporário continua apenas como referência de comparação até essa validação.

Antes de aplicar um corte, verificar se a superfície resultante ainda sustentaria árvores, rochas e construções na área atingida; se algum desses objetos perder o apoio, rejeitar o corte inteiro e informar o motivo ao jogador. A regra considera a superfície de apoio, não apenas a projeção X/Z do objeto, para permitir escavações próximas e galerias profundas que não causem objetos flutuantes. Grama e arbustos sobre uma área efetivamente escavada são removidos; seus IDs devem permanecer removidos após salvar, descarregar e recarregar regiões, inclusive nas divisas.

## Controle de desempenho

- **Malha:** retirar o OBJ temporário do caminho de carregamento e escavação depois de comprovar, num protótipo pequeno, que a criação e a liberação de uma malha em memória funcionam sem vazamentos ou acesso inválido. Na versão instalada de raylib-go, `UploadMesh` reconhece buffers geridos por Go em `UnloadMesh`, mas `UnloadModel` chama diretamente a função C que também libera as malhas. Não combinar essas chamadas sem definir e verificar a propriedade dos buffers. Continuar usando Polyform para extrair a superfície.
- **Carregamento:** manter uma quantidade limitada de regiões ativas, antecipar a próxima faixa durante o movimento e descarregar regiões distantes com margem para evitar alternância repetida. Preparar densidades e geometria em um número limitado de tarefas de CPU; enfileirar a transferência para GPU na thread de desenho, com limite de trabalho por quadro. Descartar resultados de tarefas que deixaram de ser necessárias.
- **Consultas:** raycasts, colisões, posicionamento e IA consultam apenas regiões e objetos próximos, usando o índice espacial das instâncias carregadas. O desenho considera apenas regiões visíveis. Uma escavação reconstrói só as regiões afetadas e suas bordas; recursos compartilhados, como shader e texturas, não são duplicados para cada região.
- **Texturas e mapa:** começar com poucas camadas e medir o custo do shader antes de aumentar as amostragens; se a projeção triplanar custar demais, simplificar a projeção mantendo o alinhamento nas bordas. Atualizar tiles do minimapa quando seus dados mudarem e reutilizá-los nos quadros seguintes.
- **Armazenamento:** carregar alterações por região e agrupar gravações. Nem o arquivo de save nem um futuro banco de dados devem ser consultados por objeto ou por quadro.

Registrar antes e depois de cada etapa: tempo de quadro (mediana, percentil 95 e pior caso), tempo de geração da densidade, extração da malha e envio à GPU, número de regiões/triângulos ativos e memória de CPU/GPU. Usar a mesma rota de ida e volta entre regiões, com e sem escavação, na máquina do autor. A meta inicial é manter os 60 FPS já observados no mapa atual e atravessar uma borda sem pausa perceptível; o orçamento de cerca de 16,7 ms por quadro é um alvo de medição, não uma garantia. Se a travessia ultrapassar esse orçamento, reduzir trabalho simultâneo, distância ativa ou custo de malha/material com base na etapa medida. Não avançar do checkpoint de carregamento dinâmico enquanto houver pausas reproduzíveis ou crescimento contínuo de memória nessa rota.

## Material e aparência

Usar texturas pequenas e repetíveis de grama e terra/rocha, escolhidas ou produzidas com origem e licença registradas. As imagens de prévia dos kits atuais não serão presumidas como texturas de terreno. Um material/shader compartilhado por todas as regiões usa coordenadas de mundo, para manter escala e alinhamento visual nas bordas. Grama domina superfícies voltadas para cima; terra/rocha aparece em encostas, cortes e faces inferiores. Projeção triplanar ou uma composição equivalente evita esticar a textura nas faces verticais. Normais coerentes entre regiões e iluminação simples fazem parte do mesmo checkpoint visual. A quantidade de amostragens do shader deve ser medida antes de acrescentar novas camadas.

Adicionar um céu estilizado ao marco de aparência para substituir o fundo de cor única. O céu deve manter horizonte e iluminação coerentes com o terreno, sem depender das regiões carregadas ou revelar suas divisas. A técnica visual será escolhida ao implementar, considerando qualidade e custo de renderização.

## Estradas e assentamentos

Acima das regiões de malha haverá uma camada de **planos territoriais**. Uma célula de planejamento, maior que uma região de terreno, decide de forma reproduzível locais candidatos a assentamento e conexões de estrada. Conexões em uma fronteira são determinadas por uma chave compartilhada da fronteira, para que vizinhos concordem sem exigir que todo o mundo seja gerado. Planos têm ID estável, área ocupada e geometria lógica; uma região consulta os planos que cruzam seu volume.

O traçado de uma estrada é decidido antes da composição visual local. Cada região adapta seu trecho de terreno, material e ocupação, sem gerar uma estrada independente e desconectada. Assentamentos reservam uma área que pode atravessar várias regiões e montam edifícios com peças e assets reutilizáveis. A busca de rotas em grafos pode usar Gonum quando existir uma rede que justifique essa dependência; regras de localização, relevo e estilo permanecem próprias do jogo. Recast/Detour é candidato para navegação de NPCs, não para gerar estradas ou cidades.

## Minimapa

O minimapa é composto por pequenos tiles associados às coordenadas das regiões. Cada tile resume superfície, material, estradas e pontos de interesse; tiles próximos podem ser desenhados com recursos da raylib e guardados em cache. O minimapa em tempo real não renderiza todas as regiões do mundo. Exploração descoberta pelo jogador é persistida à parte da geração. Para cavernas, uma camada ou representação específica será decidida quando houver exploração subterrânea jogável.

## Save, banco de dados e conteúdo de IA

Persistir metadados do mundo (semente, parâmetros e versão do gerador) e **alterações**: amostras escavadas, recursos removidos, construções, entidades relevantes, exploração e planos territoriais adicionados. IDs e coordenadas globais não dependem dos índices de slices ou da presença da região na memória. O save v4 da região inicial deve ser aceito e convertido para a nova representação; a nova escrita recebe versão própria.

Uma fronteira de armazenamento oferece operações por região e por plano territorial. A primeira implementação pode continuar em arquivos. Um banco de dados futuro responderá consultas por área/região quando uma região for ativada, sem consulta por objeto a cada quadro e sem guardar malhas que a semente reconstrói. Conteúdo proposto por IA entra como comandos ou planos estruturados, passa pelas mesmas validações de espaço, custo e consistência usadas pelo jogo e só então é persistido. O agente não altera diretamente GPU, malha ou tabelas.

## Entregas verificáveis

1. **Duas regiões:** a inicial e uma vizinha compartilham borda; o jogador atravessa com terreno e colisão contínuos. Objetos iniciais e saves v4 continuam funcionando.
2. **Carregamento dinâmico:** regiões próximas aparecem antes da travessia; regiões distantes são liberadas; voltar à área produz o mesmo terreno e objetos. Medir tempo de geração, tempo por quadro e uso de memória.
3. **Escavação entre regiões e suporte:** um corte sobre a borda atualiza os dois lados e reaparece corretamente após salvar, descarregar e recarregar. Cortes que deixariam árvore, rocha ou construção sem apoio são rejeitados sem alterar o terreno; grama e arbustos atingidos por cortes válidos desaparecem e não reaparecem após o recarregamento.
4. **Aparência:** grama e solo exposto têm texturas estilizadas, iluminação legível e nenhuma emenda perceptível de textura ou normal; um céu estilizado substitui o fundo de cor única e acompanha visualmente o mundo expansível.
5. **Conteúdo territorial:** uma estrada atravessa ao menos duas regiões sem quebra; um assentamento pequeno usa assets existentes e respeita a ocupação.
6. **Minimapa:** mostra terreno e estrada ao redor do jogador, mantém posição correta durante a travessia e preserva áreas descobertas.

O primeiro guia de implementação tratará somente a entrega 1. Cada entrega será dividida em passos pequenos com trechos de código, localização exata e verificação visual pelo autor.

## Riscos a verificar

- A malha atual passa por arquivo temporário; repetir isso durante carregamento pode causar pausas. Validar a alternativa em memória e sua liberação antes de adotá-la.
- Amostras de borda iguais não bastam se normais ou texturas usarem coordenadas locais diferentes.
- Posições globais em `float32` perdem precisão longe da origem. Manter identidade global inteira e posições locais por região; adotar deslocamento de origem de renderização antes que a distância torne isso observável.
- Geração e carregamento devem ter orçamento de memória e de tempo; “mundo infinito” não implica manter todas as regiões ativas.
- Save, geração de objetos, navegação e colisão ainda contêm suposições da região inicial; a primeira travessia deve verificar cada uma.

## Bibliotecas e fontes de referência

Reutilizar FastNoiseLite, Polyform e raylib-go já presentes. Avaliar Gonum Graph quando houver rede de estradas; avaliar Recast/Detour apenas se a navegação dos personagens exigir. Nenhuma biblioteca substitui as regras específicas deste mundo.

- [raylib: modelos, shaders, materiais e renderização em textura](https://www.raylib.com/cheatsheet/cheatsheet.html)
- [Godot: descrição da projeção triplanar](https://docs.godotengine.org/en/4.3/tutorials/3d/standard_material_3d.html)
- [Gonum Graph](https://pkg.go.dev/gonum.org/v1/gonum/graph)
- [Recast Navigation](https://github.com/recastnavigation/recastnavigation)
