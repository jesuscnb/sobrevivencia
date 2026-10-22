#version 330

in vec3 worldPosition;
in vec3 worldNormal;

uniform sampler2D grassTexture;
uniform sampler2D soilTexture;
uniform vec4 colDiffuse;
uniform vec3 sunDirection;
uniform sampler2D roadMask;
uniform vec2 regionMinXZ;

out vec4 finalColor;

void main() {
    vec3 normal = normalize(worldNormal);
    vec3 absoluteNormal = abs(normal);

    // A projeção usa coordenadas globais; a textura continua na divisa.
    vec2 grassUV = worldPosition.xz * 0.32;
    vec2 soilUV;
    if (absoluteNormal.y >= absoluteNormal.x && absoluteNormal.y >= absoluteNormal.z) {
        soilUV = worldPosition.xz * 0.32;
    } else if (absoluteNormal.x >= absoluteNormal.z) {
        soilUV = worldPosition.zy * 0.32;
    } else {
        soilUV = worldPosition.xy * 0.32;
    }

    float grassWeight = smoothstep(0.45, 0.78, normal.y);
    vec3 grass = texture(grassTexture, grassUV).rgb;
    vec3 soil = texture(soilTexture, soilUV).rgb;
    vec3 surface = mix(soil, grass, grassWeight);

    // 64 pixels internos, mais um pixel de borda em cada lado.
    vec2 roadUV = (vec2(1.0) + (worldPosition.xz - regionMinXZ) * 4.0) / 66.0;

    float roadWeight = texture(roadMask, roadUV).r;
    roadWeight *= smoothstep(0.45, 0.80, normal.y);

    vec3 roadSurface = soil * vec3(1.10, 0.82, 0.57);
    surface = mix(surface, roadSurface, roadWeight);

    float diffuse = max(dot(normal, normalize(sunDirection)), 0.0);
    float light = 0.56 + 0.44 * diffuse;
    finalColor = vec4(surface * light * colDiffuse.rgb, colDiffuse.a);
}
