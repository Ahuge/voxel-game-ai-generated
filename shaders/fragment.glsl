#version 410 core

in vec3 ourColor;
in vec2 TexCoord;
in vec3 Normal;
in vec3 FragPos;

out vec4 FragColor;

uniform sampler2D texture1;

// Light properties
uniform vec3 lightPos = vec3(100.0, 100.0, 100.0); // Closer light source for better illumination
uniform vec3 lightColor = vec3(1.0, 1.0, 0.9); // Slightly yellow sunlight
uniform vec3 viewPos = vec3(0.0, 0.0, 0.0); // Will be updated to camera position

void main()
{
    // Ambient lighting
    float ambientStrength = 0.8; // Reduced ambient to allow diffuse to have more effect
    vec3 ambient = ambientStrength * lightColor;
    
    // Diffuse lighting
    vec3 norm = normalize(Normal);
    vec3 lightDir = normalize(lightPos - FragPos);
    float diff = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = diff * lightColor * 0.8; // Added multiplier for better control
    
    // Specular lighting
    float specularStrength = 0.3; // Increased for more visible highlights
    vec3 viewDir = normalize(viewPos - FragPos);
    vec3 reflectDir = reflect(-lightDir, norm);
    float spec = pow(max(dot(viewDir, reflectDir), 0.0), 16); // Reduced shininess for broader highlights
    vec3 specular = specularStrength * spec * lightColor;
    
    // Combine lighting with texture and vertex color
    vec3 lighting = ambient + diffuse + specular;
    
    // Add a minimum lighting level to prevent completely dark areas
    lighting = max(lighting, vec3(0.2));
    
    // Sample the texture
    vec4 texColor = texture(texture1, TexCoord);
    
    // Apply lighting to the texture, preserving the texture's appearance better
    FragColor = texColor * vec4(lighting * ourColor, 1.0);
}