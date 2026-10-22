# Novo projeto: jogo 3D de sobrevivência

**Estado:** plano de transferência atualizado em 17/10/2026. Este arquivo pode ser copiado para o novo repositório. Go foi escolhido para cliente e futuro servidor; as bibliotecas específicas ainda serão validadas por etapa.

## 1. Decisões do autor

- Criar um **jogo real, simples e jogável** para aprender desenvolvimento de jogos.
- **Tratar o projeto como um jogo destinado a lançamento**, mesmo sendo um estudo. Em cada etapa, comparar alternativas e escolher uma solução adequada ao produto final. Protótipos visuais e atalhos de aprendizado devem ser identificados como temporários, com o caminho de substituição claro; não presumir que a técnica já estudada é a melhor.
- Trabalhar **sem engine** como Unity, Unreal ou Godot, mas usar bibliotecas por responsabilidade para evitar reimplementar toda a matemática e infraestrutura.
- O **autor escreve o código**. O assistente orienta, pesquisa, compara, responde dúvidas, propõe exercícios, revisa e ajuda a depurar. Só deve gerar código, assets ou funcionalidades completas mediante pedido específico.
- O início será **3D em primeira pessoa** e para **Windows PC**.
- O mundo começa pelo terreno procedural; depois vêm vegetação, rochas e objetos, construções, inimigos, personagem, interfaces e, por último, multiplayer.
- **Escavação 3D livre está confirmada como recurso futuro:** será possível abrir túneis e cavernas por baixo do chão. A ferramenta de escavação pode vir depois da primeira versão jogável, mas os dados do terreno devem permitir isso desde o começo.
- Construções em duas etapas: primeiro **peças colocadas pelo jogador**; depois **estruturas geradas automaticamente pelo mundo**.
- **Linguagem escolhida: Go para o cliente e para o futuro servidor.** O multiplayer continua para depois do jogo solo.
- Os assets podem seguir um estilo **low poly**. O **terreno precisa ser procedural**; a técnica de geração ainda será escolhida com orientação.
- Exploração, abrigo, inimigos, coleta e fabricação servem como contexto inicial de sobrevivência. **A jogabilidade detalhada não é a prioridade da pesquisa agora.**
- As primeiras estruturas geradas pelo mundo serão **ruínas e cabanas**; povoações ficam para uma evolução.
- Go, Java, Python e Rust interessam como linguagens modernas. Avaliar também o uso de bibliotecas nativas C/C++ a partir delas, sem presumir que essa integração será gratuita.
- Multiplayer não tem objetivo comercial neste projeto de aprendizado. Número de jogadores e hospedagem serão decididos apenas quando o jogo solo existir.

## 2. O que levar do VolumetricLab

O projeto atual é um laboratório C++20 de geometria e sistemas de mundo, não a base obrigatória do novo jogo. Os experimentos estudaram ruído, terreno, SDF (campo de distância), chunks (partes do mundo), reconstrução localizada, níveis de detalhe, transições de malha, materiais, luz e simulações de rede. No novo jogo, a lição é reconhecer o problema e saber escolher uma biblioteca ou uma solução simples, sem refazer obrigatoriamente as implementações didáticas.

O design antigo previa construir a tecnologia volumétrica manualmente. **Essa regra não vale para o novo projeto.** O mini projeto de escultor criado no repositório atual foi uma interpretação anterior do pedido; não define o rumo deste jogo. Este plano não altera nem reaproveita automaticamente seu código.

Preferências de ensino já expressas: visão geral antes da matemática, explicação simples, poucos passos grandes com resultado visível, siglas explicadas ao aparecerem, comentários didáticos quando houver código solicitado e chaves explícitas nas estruturas de controle, em estilo próximo ao Java. Não transformar cada variação matemática em lição separada.

## 3. Primeiro jogo completo: sugestão de recorte

Um mapa pequeno gerado por uma **semente**. O jogador caminha em primeira pessoa, encontra árvores e rochas, coleta madeira e pedra, constrói um abrigo simples, enfrenta um tipo de inimigo, acompanha vida e inventário e consegue salvar e retomar a partida. Esse ciclo permite jogar, mas a mecânica exata e uma condição de vitória podem ser definidas mais tarde. Mundo infinito, dezenas de biomas e multiplayer não são requisitos da primeira versão.

“Procedural” pode significar coisas diferentes:

1. **Terreno:** uma semente e regras geram a forma inicial e tipos de solo. Como haverá escavação, estudar um campo volumétrico desde o início, mesmo que a primeira superfície pareça um terreno de alturas.
2. **Vegetação e rochas:** regras escolhem **onde** colocar modelos; esses modelos podem ser assets prontos. Gerar posições não é o mesmo que gerar a geometria de cada árvore.
3. **Construção do jogador:** regras de encaixe, rotação, custo e suporte combinam peças.
4. **Estruturas do mundo:** mais tarde, regras montam ruínas e cabanas com peças modulares; vilas podem vir depois.

Uma mesma semente deve reconstruir o mesmo cenário. Mudanças do jogador, como árvores coletadas e paredes construídas, devem ser registradas à parte da geração. Esta é uma proposta de arquitetura a validar.

### Perlin Noise e Transvoxel: funções diferentes

**Perlin Noise** é uma opção de ruído: para cada posição horizontal, a biblioteca devolve um valor que podemos converter em altura. Podemos combinar escalas de ruído e regras de bioma, sem implementar a fórmula do ruído. [FastNoiseLite](https://github.com/Auburn/FastNoiseLite) já oferece Perlin e outros ruídos em C++, Java, Go e Rust. **O ruído não desenha o terreno sozinho**: ainda escolhemos a regra de altura, criamos a malha e a colisão e distribuímos vegetação.

**Transvoxel** resolve outro problema: unir malhas de **terreno volumétrico** geradas em resoluções diferentes, para evitar rachaduras nas bordas entre níveis de detalhe. Ele não gera ruído nem substitui a geração de terreno. O [material original do algoritmo](https://transvoxel.org/) fornece explicação e tabelas, mas não é, por si só, uma biblioteca completa de mundo pronta para qualquer linguagem.

**Decisão de arquitetura para a escavação 3D livre:** usar um **campo de densidade 3D desde o primeiro terreno**. No começo, a regra pode continuar simples: uma altura vinda de ruído 2D define onde existe solo, e um gerador de malha volumétrica transforma a fronteira em triângulos. Visualmente parecerá um terreno comum, mas os dados poderão receber cortes e cavernas mais tarde. Isso custa mais memória e integração agora do que uma malha de alturas pura; evita trocar a representação central quando chegar a escavação. **Transvoxel ainda não precisa entrar no primeiro terreno**: um único nível de detalhe basta para validar geração, desenho e uma edição local. Ele entra quando regiões volumétricas vizinhas passarem a ter resoluções diferentes.

## 4. Marcos grandes, cada um com resultado jogável

| Marco | Resultado observável | Conceito a praticar | Biblioteca candidata somente quando necessária |
|---|---|---|---|
| 0. Base | Janela 3D e câmera de teste no novo projeto, feitas pelo autor | Atualização, desenho, entrada, coordenadas e build | Avaliar raylib-go como primeira biblioteca de janela e desenho |
| 1. Terreno | Pequeno volume reproduzível pela semente, com superfície suave | Ruído define solo; densidade 3D vira malha e colisão | FastNoiseLite + mesher volumétrico a validar em Go |
| 2. Jogador e prova de edição | Caminhar, olhar, colidir e fazer um corte de teste no volume | Separar entrada, movimento, câmera, dados do terreno e atualização da malha | Colisão simples/física da stack; mesher escolhido |
| 3. Natureza e objetos | Grama, árvores e rochas distribuídas por regras | Distribuição, inclinação, distância, reuso de modelos | Carregamento de modelos; desenho em lotes quando necessário |
| 4. Sobrevivência | Coletar madeira/pedra, inventário mínimo e vida | Interação pela mira, estado do jogo e feedback | UI da stack; formato simples de save |
| 5. Construção | Posicionar piso, parede e teto; salvar e recarregar | Prévia, encaixe, custo, validação e peças modulares | Raycast e colisão da stack escolhida |
| 6. Ameaça | Um inimigo patrulha, persegue e ataca | Estados de IA, dano, balanceamento e animação básica | Navegação simples primeiro; biblioteca se o mapa pedir |
| 7. Escavação e expansão | Escavar no jogo, persistir cortes e respeitar o apoio dos objetos; ruínas/cabanas geradas, som, céu e mais regiões | Edição volumétrica local, colisão atualizada, composição procedural, aparência do ambiente e desempenho | Mesher/colisão; Transvoxel apenas se houver níveis de detalhe diferentes |
| 8. Multiplayer, se chegar a hora | Jogadores compartilham o mundo solo já estável; quantidade e hospedagem ainda abertas | Estado autoritativo, sincronização, latência e persistência | Biblioteca de rede escolhida nessa fase |

O **jogador mínimo aparece cedo** porque caminhar pelo mapa ajuda a verificar o terreno. Animações, habilidades e refinamento do personagem continuam como assunto próprio mais adiante. Nenhum marco exige muitos experimentos quase iguais.

**Suporte após escavação no marco 7:** impedir um corte que retiraria a superfície de apoio de árvores, rochas ou construções, antes de alterar o terreno. Cortes próximos ou profundos que preservem esse apoio continuam permitidos. Grama e arbustos atingidos por um corte válido são removidos, e essa remoção deve sobreviver ao save e ao descarregamento/recarregamento da região.

**Teste de decisão criado no laboratório atual a pedido do autor:** os protótipos em mini_projects/go_rust_terrain_compare geram um volume 3D com a mesma semente e dimensões equivalentes. Cada um mostra a superfície e aplica um corte esférico de demonstração. Para isolar a comparação das linguagens e das bibliotecas de malha, a forma inicial usa ondas simples (sin/cos), ainda não Perlin Noise. Critérios de comparação: facilidade de compilar, clareza do caminho dos dados (densidade → malha → tela), quantidade de integração necessária e facilidade de alterar o corte. A jogabilidade da ferramenta de escavação pode vir depois. No novo repositório, o autor pode estudar esses protótipos e reconstruir conscientemente apenas o caminho escolhido.

## 5. Linguagens e bibliotecas: comparação para o projeto inteiro

Bibliotecas/frameworks de programação são compatíveis com “sem engine” no sentido usado aqui: fornecem APIs, sem editor de cenas e fluxo completo de produção. **Nenhuma das opções entrega sozinha um jogo de sobrevivência pronto.** Todas permitem terreno procedural; a diferença é quanto trabalho resta para integrar física, modelos, interface e rede. A avaliação abaixo é uma **recomendação de projeto**, não um benchmark de desempenho.

### Caminhos considerados para o cliente 3D

| Caminho | Bibliotecas iniciais plausíveis | Pontos fortes | Atenção para este projeto |
|---|---|---|---|
| **Java + libGDX** | [libGDX](https://libgdx.com/features/) para 3D, entrada, áudio e UI; [FastNoiseLite Java](https://github.com/Auburn/FastNoiseLite); extensão [Bullet](https://libgdx.com/wiki/extensions/physics/bullet/bullet-physics) quando necessária | Conjunto relativamente integrado; sintaxe próxima à convenção solicitada; há [API de malhas](https://libgdx.com/wiki/graphics/opengl-utils/meshes), [animação](https://libgdx.com/wiki/graphics/3d/3d-animations-and-skinning) e interface para os marcos futuros. A extensão Bullet já chama código C++ por trás do Java. | O autor ainda monta o terreno e os sistemas de jogo. O formato glTF/.glb exige uma extensão como [gdx-gltf](https://github.com/mgsx-dev/gdx-gltf), cuja compatibilidade deve ser provada com um asset real. |
| **Rust + raylib-rs** | [raylib-rs](https://github.com/raylib-rs/raylib-rs), [FastNoiseLite Rust](https://github.com/Auburn/FastNoiseLite), [transvoxel_rs](https://github.com/gnurfos/transvoxel_rs) para extrair a malha, [Rapier](https://rapier.rs/) quando a física for necessária | Bibliotecas Rust para desenho, ruído, malha volumétrica com transições de nível de detalhe e física. A biblioteca de malha também funciona com um nível de detalhe só. | Aprende Rust e desenvolvimento 3D ao mesmo tempo. A biblioteca transvoxel_rs gera a malha de uma região, mas não gerencia mundo, edições, colisão ou vizinhos; sua adequação e manutenção precisam ser verificadas no teste. |
| **Go + raylib-go** | [raylib-go](https://github.com/gen2brain/raylib-go), [FastNoiseLite Go](https://github.com/Auburn/FastNoiseLite), [Polyform/marching](https://pkg.go.dev/github.com/EliCDavis/polyform/modeling/marching) como candidato para extrair malha | Sintaxe pequena; raylib cobre 3D, modelos, áudio e entrada. Polyform expõe campo 3D, Marching Cubes e subtração de formas; há caminho com cgo e caminho sem cgo para raylib no Windows. | É preciso testar a conversão da malha de Polyform para raylib, reconstrução após edição e colisão. Não há transições de Transvoxel documentadas nesse candidato; se níveis de detalhe forem necessários, avaliar outra solução ou integração nativa. |
| **Python + raylib-python-cffi** | [binding raylib em Python](https://github.com/electronstudio/raylib-python-cffi); biblioteca de ruído/NumPy a avaliar | Prototipagem rápida, código curto e acesso à raylib nativa. [PyBullet](https://github.com/bulletphysics/bullet3/blob/master/docs/pybullet_quickstart_guide/PyBulletQuickstartGuide.md.html) existe para física. | Loops Python sobre muitos pontos de terreno/vegetação e chamadas nativas muito pequenas podem exigir processamento em lote ou módulos nativos. Integrar física, renderer e distribuição do executável adiciona trabalho. Melhor candidato para ferramentas e protótipos do que para o cliente principal deste estudo. |
| **C++20 + raylib** | [raylib](https://github.com/raysan5/raylib/blob/master/README.md), [FastNoiseLite C++](https://github.com/Auburn/FastNoiseLite), [Jolt Physics](https://github.com/jrouwe/JoltPhysics) quando necessário | Acesso direto às bibliotecas nativas; reaproveita o aprendizado do laboratório; opção forte se o terreno virar volumétrico e precisar de otimização fina. | Mais gerenciamento manual de recursos, build e integração. Pode desviar o tempo de estudo do jogo para infraestrutura. |

**C#** foi considerado como alternativa: [MonoGame](https://docs.monogame.net/articles/getting_to_know/whatis/graphics/WhatIs_3DRendering.html) tem pipeline de conteúdo, mas pede mais montagem do 3D; o binding [raylib-cs](https://github.com/raylib-cs/raylib-cs) informa manutenção passiva. Não é necessário continuar comparando linguagens após a escolha de Go.

**Decisão após os protótipos:** usar **Go no cliente**. raylib-go é o primeiro candidato para janela e desenho, pois funcionou no teste, mas a biblioteca principal não foi congelada. Rust foi comparado, mas não é uma segunda implementação a manter. O autor conhece a sintaxe, o coletor de lixo e a concorrência de Go. Para recursos 3D que uma biblioteca Go não atender bem, avaliar uma biblioteca C/C++ existente antes de implementar a matemática manualmente. Essa possibilidade não obriga o projeto a ter um núcleo C++ desde o início. Os protótipos não mediram desempenho de um mundo grande nem provaram a integração com física 3D.

### Cobertura das etapas, do terreno ao multiplayer

Legenda: **direto** = existe recurso documentado na stack indicada; **integrar** = juntar bibliotecas ou escrever uma camada própria; **regras do jogo** = o trabalho é semelhante em qualquer linguagem.

| Etapa | Java/libGDX | Rust/raylib-rs | Go/raylib-go | Python/raylib-CFFI | C++/raylib |
|---|---|---|---|---|---|
| Terreno 3D escavável | FastNoiseLite + mesher volumétrico a escolher + libGDX | FastNoiseLite + transvoxel_rs + conversão para malha raylib | FastNoiseLite + Polyform/marching + conversão para malha raylib | Ruído/arrays em lote + mesher volumétrico a escolher | FastNoiseLite + mesher volumétrico a escolher |
| Transições entre resoluções, se necessárias | Biblioteca/integração a avaliar | transvoxel_rs é candidato; mundo e vizinhos continuam por nossa conta | Biblioteca/integração a avaliar | Biblioteca/integração a avaliar | Biblioteca/integração a avaliar |
| Vegetação e rochas | Modelos, instâncias e cache; otimizar se preciso | Modelos raylib; otimizar se preciso | Modelos raylib; otimizar se preciso | Modelos raylib; evitar muitos loops Python | Modelos raylib; otimizar se preciso |
| Construção e ruínas | Regras do jogo + modelos/colisão | Regras do jogo + modelos/colisão | Regras do jogo + modelos/colisão | Regras do jogo + modelos/colisão | Regras do jogo + modelos/colisão |
| Jogador e colisão | Bullet já tem wrapper oficial | Rapier é opção própria | Colisão simples raylib; física rica pede integração | Colisão simples raylib; PyBullet pede integração | Colisão simples raylib; Jolt pode ser integrado |
| Monstros, caminho | Estados/grade; Recast depois se necessário | Estados/grade; Recast depois se necessário | Estados/grade; Recast depois se necessário | Estados/grade; atenção ao custo de muitos agentes | Estados/grade; Recast depois se necessário |
| Modelos e animação | API 3D; glTF requer extensão | Binding raylib para modelos/animação | Binding raylib para modelos/animação | Binding raylib para modelos/animação | API raylib para modelos/animação |
| Vida, inventário, menus | scene2d.ui | raygui ou UI própria | raygui ou UI própria | raygui ou UI própria | raygui ou UI própria |
| Save e multiplayer tardio | Serialização/rede Java; protocolo próprio | serde/rede Rust; protocolo próprio | JSON/rede Go; protocolo próprio | JSON/rede Python; protocolo próprio | Serialização/rede C++; protocolo próprio |

Esta tabela mostra **disponibilidade de blocos**, não que tudo se encaixe automaticamente. A biblioteca de navegação [Recast/Detour](https://github.com/recastnavigation/recastnavigation) é C++; se for realmente necessária em outra linguagem, haverá uma etapa de integração nativa. Um caminho em grade pode bastar para o primeiro inimigo. As fontes de cada stack estão ligadas na tabela anterior; para vegetação em libGDX, ver também [ModelCache](https://libgdx.com/wiki/graphics/3d/modelcache).

### Invocar C/C++ a partir de outra linguagem

**Sim, é possível.** Em Go, a rota inicial para uma biblioteca nativa com uma interface **C** pequena é [cgo](https://pkg.go.dev/cmd/cgo). Para bibliotecas que expõem classes C++, [SWIG](https://www.swig.org/Doc4.5/Go.html) é outra opção a avaliar; também é possível criar uma pequena camada C sobre a biblioteca C++. Escolher a rota ao conhecer a API concreta, sem presumir que a integração será automática.

O fluxo conceitual mais simples é: **Go → binding → API C → biblioteca C++**. Para uma operação pesada, chamar uma função por região ou lote de dados, como "reconstruir estas regiões", em vez de uma chamada por vértice. Definir quem aloca e libera a memória nativa, como os erros retornam, formatos de arrays e como distribuir a DLL faz parte dessa integração. O coletor de lixo de Go não libera automaticamente objetos criados em C++. Usar primeiro bibliotecas e bindings existentes; criar uma ponte própria quando uma necessidade concreta justificar.

### Cliente e servidor

**Cliente e futuro servidor usarão Go.** Isso permite compartilhar alguns tipos e regras quando fizer sentido, mantendo cliente e servidor como programas separados. O servidor futuro deve validar ações e controlar o estado persistente. Transporte, hospedagem e protocolo serão decididos somente depois do jogo solo; [ENet](https://github.com/lsalzman/enet) é apenas um candidato de transporte para avaliar nessa fase.

### O que os protótipos confirmaram

Os dois protótipos no laboratório atual usam um pequeno terreno criado da mesma semente em um campo 3D, desenhado a partir de uma malha volumétrica, e um corte esférico que muda essa malha. Eles mostram um caminho viável em Go e Rust, mas **não** medem desempenho de mundo grande nem testam colisão. O protótipo Go é referência de estudo, não base obrigatória do jogo. Antes de depender de C++ ou fechar a solução de física, validar uma pequena região editável com colisão atualizada.

## 6. Bibliotecas por problema, não por antecipação

| Problema | Candidato | O que ele resolve; o que fica conosco |
|---|---|---|
| Ruído procedural | [FastNoiseLite](https://github.com/Auburn/FastNoiseLite) tem versões em C++, Java, Go e Rust; Python requer biblioteca ou binding adicional | Produz números de ruído 2D/3D, incluindo Perlin. Decidir como viram altura, bioma, posição de árvores e semente ainda é trabalho do jogo. **Não gera sozinho um mundo jogável.** |
| Superfície do terreno escavável | [Polyform/marching](https://pkg.go.dev/github.com/EliCDavis/polyform/modeling/marching) em Go; outra biblioteca, inclusive nativa, se surgirem limites concretos | Extrai triângulos de um campo 3D. Organização das regiões, persistência, atualização local, colisão e eventuais transições de nível de detalhe continuam sendo responsabilidades a resolver no jogo. |
| Colisão e movimento | Colisão simples para começar; avaliar física 3D existente, inclusive uma biblioteca C++ como [Jolt Physics](https://github.com/jrouwe/JoltPhysics), quando necessária | Física resolve colisões/corpos; controles e sensação do personagem são decisões nossas. Verificar o binding Go e a atualização da colisão após escavação antes de adotar uma biblioteca nativa. |
| Caminho dos inimigos | Estados simples e caminho por grade; depois [Recast/Detour](https://github.com/recastnavigation/recastnavigation) se necessário | Navegação encontra um trajeto; perceber o jogador, atacar e reagir às construções continuam sendo IA do jogo. |
| Interface | UI da stack; com raylib, avaliar [raygui](https://github.com/raysan5/raygui) | Barras, inventário e menus; adiar a escolha visual até existir interação real. |
| Modelos/animação | Importador da stack e formato compatível; glTF/.glb como candidato de intercâmbio | O formato pode transportar modelos e animações, mas é preciso verificar importação, escala e animação **na stack escolhida**. [Fonte: Khronos glTF](https://www.khronos.org/gltf/). |

Começar com a solução simples, observar o limite e só então integrar física, navegação ou rede dedicada.

## 7. Assets e direção visual

**Decisão atual:** usar assets low poly. Começar com formas e cores simples para validar o ciclo; depois trocar por um conjunto visual coerente. Vegetação procedural não obriga a modelar cada árvore por algoritmo: modelos prontos podem ser distribuídos por regras determinísticas, com variações moderadas de escala e rotação.

- O [Kenney Survival Kit](https://kenney.nl/assets/survival-kit) é um exemplo de conjunto coerente de peças 3D. A [página de suporte da Kenney](https://kenney.nl/support) informa as condições CC0 dos assets que ela oferece.
- [Poly Haven](https://polyhaven.com/license) publica assets sob CC0 e pode ajudar com texturas, modelos e iluminação de ambiente.
- Registrar **origem, licença, link e alterações** de cada pacote importado, mesmo quando atribuição não for obrigatória.
- Escolher ferramenta de modelagem/animação quando houver necessidade concreta. Se for usado Blender, testar a exportação que a stack consegue ler antes de criar muitos modelos.
- Consistência de poucos assets importa mais no primeiro jogo do que grande variedade de árvores, rochas e monstros.

## 8. Decisões técnicas que mudam a arquitetura

1. **Volume desde o começo, escavação jogável depois.** A escavação 3D livre exige representar matéria e ar inclusive debaixo da superfície; uma única altura por posição horizontal não basta. O primeiro mapa pode parecer um relevo comum, mas sua fonte de dados será 3D. Primeiro provar um corte pequeno; depois acrescentar ferramenta jogável, atualização local da malha, colisão e save. Transvoxel só entra se houver regiões vizinhas com níveis de detalhe diferentes.
2. **Semente e alterações separadas.** Salvar semente, versão do gerador, personagem, inventário, recursos removidos, construções e, quando existirem, cortes no terreno. Mudanças futuras no gerador podem exigir migração de saves; definir formato no primeiro save.
3. **Unidade, eixos e escala.** Combinar no começo o que representa um metro e qual eixo aponta para cima para compatibilizar chão, colisão e modelos.
4. **Mundo limitado primeiro.** Criar uma região pequena de densidade 3D. Dividir em chunks (regiões do mundo), carregar sob demanda e adicionar níveis de detalhe quando houver necessidade observável. Planejar coordenadas globais desde o início para que as regiões vizinhas amostrem a mesma matéria nas bordas.
5. **Regras separadas do desenho.** Vida, inventário, coleta e construção não devem depender diretamente das chamadas que desenham modelos. Isso ajuda no save e em um futuro servidor.
6. **Configurações do mundo e do jogo.** Manter um ponto central para os parâmetros realmente usados na geração do mundo e para preferências como áudio, vídeo e controles. Planejar uma interface dentro do jogo para consultar e alterar as opções pertinentes quando a UI for desenvolvida. Salvar com cada partida os parâmetros e a versão do gerador que definem seu mundo; mudanças nesses valores podem exigir um novo mundo ou uma migração explícita. Preferências gerais podem ser alteradas sem recriar a partida. Definir as opções aos poucos, conforme surgirem necessidades concretas.

## 9. Forma de colaboração no novo projeto

Para cada marco, apresentar **(1) resultado visível, (2) conceito em linguagem comum, (3) biblioteca e duas ou três APIs relevantes, (4) tarefa para o autor implementar, (5) verificação simples**. Explicar siglas na primeira aparição. Se código for solicitado, usar nomes claros, chaves nos condicionais e laços, e comentários que expliquem o fluxo sem comentar linhas óbvias. Não entregar um sistema pronto quando a intenção for praticar.

Pesquisar documentação **oficial e atual** antes de recomendar versões, APIs, compatibilidade, licenças ou dependências. Distinguir fato verificado de sugestão e apresentar bibliotecas na fase em que serão úteis.

## 10. Próximas decisões, sem pressa

1. **Agora:** iniciar o novo projeto em Go, com uma janela 3D e câmera simples. Usar o protótipo Go como referência para o campo de densidade e o corte, sem copiar um sistema pronto. Validar a colisão sobre uma pequena região editável antes de escolher física nativa.
2. **Após terreno, movimento e construção básica:** decidir uma meta curta de sobrevivência e a condição de progresso do jogo.
3. **Após o jogo solo estar estável:** decidir se haverá multiplayer no projeto de aprendizado e, nesse caso, quantos jogadores e como hospedar. Se essa fase acontecer, o servidor será em Go.

**Ao abrir o novo projeto:** copie este arquivo para a pasta docs, leia primeiro as seções 1, 3 e 5 e use o protótipo Go do laboratório atual como referência de estudo. O restante é roteiro revisável, não especificação congelada. O assistente só deve escrever novos sistemas ou código mediante pedido específico do autor.
