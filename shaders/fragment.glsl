#version 410 core

in vec3 ourColor;
in vec2 TexCoord;
in vec3 Normal;
in vec3 FragPos;

out vec4 FragColor;

uniform sampler2D texture1;

// Light properties
uniform vec3 lightPos = vec3(1000.0, 1000.0, 1000.0); // Distant light source (sun)
uniform vec3 lightColor = vec3(1.0, 1.0, 0.9); // Slightly yellow sunlight
uniform vec3 viewPos = vec3(0.0, 0.0, 0.0); // Will be updated to camera position

void main()
{
    // Ambient lighting
    float ambientStrength = 0.3;
    vec3 ambient = ambientStrength * lightColor;
    
    // Diffuse lighting
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(lightPos - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * lightColor;
    
    // Specular lighting
    float specularStrength = 0.1;
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 32);
    vec3 specular = specularStrength * spec * lightColor;
    
    // Combine lighting with texture and vertex color
    vec3 result = (ambient + diffuse + specular) * ourColor;
    
    // Apply texture if available, otherwise use the calculated color
    FragColor = texture(texture1, TexCoord) * vec4(result, 1.0);
}