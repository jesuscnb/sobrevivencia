#version 330

uniform vec2 resolution;
uniform vec3 cameraForward;
uniform vec3 cameraRight;
uniform vec3 cameraUp;
uniform float projectionScale;
uniform vec3 sunDirection;

out vec4 finalColor;

void main() {
    vec2 screen = 2.0 * gl_FragCoord.xy / resolution - 1.0;
    float aspect = resolution.x / resolution.y;
    vec3 direction = normalize(cameraForward
        + cameraRight * screen.x * aspect * projectionScale
        + cameraUp * screen.y * projectionScale);

    vec3 horizon = vec3(0.56, 0.72, 0.74);
    vec3 zenith = vec3(0.18, 0.36, 0.59);
    vec3 belowHorizon = vec3(0.57, 0.63, 0.59);
    float upperSky = smoothstep(0.0, 0.72, direction.y);
    float lowerSky = 1.0 - smoothstep(-0.30, 0.0, direction.y);
    vec3 color = mix(horizon, zenith, upperSky);
    color = mix(color, belowHorizon, lowerSky);

    float sunFacing = max(dot(direction, normalize(sunDirection)), 0.0);
    float glow = pow(sunFacing, 32.0) * 0.12;
    float disc = smoothstep(0.9992, 0.9998, sunFacing);
    color += glow * vec3(1.0, 0.82, 0.56);
    color = mix(color, vec3(1.0, 0.92, 0.71), disc);

    finalColor = vec4(color, 1.0);
}
