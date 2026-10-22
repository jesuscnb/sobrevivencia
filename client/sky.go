package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Ceu desenha apenas o fundo. Ele não depende das regiões carregadas do mundo.
type Ceu struct {
	shader         rl.Shader
	localResolucao int32
	localFrente    int32
	localDireita   int32
	localAcima     int32
	localEscala    int32
}

func carregarCeu() Ceu {
	ceu := Ceu{shader: rl.LoadShader("", "assets/sky/sky.fs")}
	if !rl.IsShaderValid(ceu.shader) {
		panic("não foi possível carregar o shader do céu")
	}
	ceu.localResolucao = rl.GetShaderLocation(ceu.shader, "resolution")
	ceu.localFrente = rl.GetShaderLocation(ceu.shader, "cameraForward")
	ceu.localDireita = rl.GetShaderLocation(ceu.shader, "cameraRight")
	ceu.localAcima = rl.GetShaderLocation(ceu.shader, "cameraUp")
	ceu.localEscala = rl.GetShaderLocation(ceu.shader, "projectionScale")
	localSol := rl.GetShaderLocation(ceu.shader, "sunDirection")
	if ceu.localResolucao < 0 || ceu.localFrente < 0 || ceu.localDireita < 0 ||
		ceu.localAcima < 0 || ceu.localEscala < 0 || localSol < 0 {
		rl.UnloadShader(ceu.shader)
		panic("shader do céu sem parâmetros da câmera e do sol")
	}
	rl.SetShaderValue(ceu.shader, localSol, direcaoSol[:], rl.ShaderUniformVec3)
	return ceu
}

func (ceu *Ceu) desenhar(camera rl.Camera3D) {
	largura := rl.GetRenderWidth()
	altura := rl.GetRenderHeight()
	if largura <= 0 || altura <= 0 {
		return
	}
	frente := rl.Vector3Normalize(rl.Vector3Subtract(camera.Target, camera.Position))
	direita := rl.Vector3CrossProduct(frente, camera.Up)
	if rl.Vector3Length(direita) < 0.0001 {
		direita = rl.Vector3CrossProduct(frente, rl.NewVector3(0, 0, 1))
	}
	direita = rl.Vector3Normalize(direita)
	acima := rl.Vector3Normalize(rl.Vector3CrossProduct(direita, frente))
	escala := float32(math.Tan(float64(camera.Fovy) * math.Pi / 360))

	rl.SetShaderValue(ceu.shader, ceu.localResolucao,
		[]float32{float32(largura), float32(altura)}, rl.ShaderUniformVec2)
	rl.SetShaderValue(ceu.shader, ceu.localFrente,
		[]float32{frente.X, frente.Y, frente.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(ceu.shader, ceu.localDireita,
		[]float32{direita.X, direita.Y, direita.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(ceu.shader, ceu.localAcima,
		[]float32{acima.X, acima.Y, acima.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(ceu.shader, ceu.localEscala,
		[]float32{escala}, rl.ShaderUniformFloat)

	rl.BeginShaderMode(ceu.shader)
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.White)
	rl.EndShaderMode()
}

func (ceu *Ceu) descarregar() {
	if ceu.shader.ID != 0 {
		rl.UnloadShader(ceu.shader)
		ceu.shader = rl.Shader{}
	}
}
