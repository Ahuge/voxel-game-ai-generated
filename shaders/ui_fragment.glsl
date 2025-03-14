#version 410 core

in vec4 ourColor;
in vec2 TexCoord;

out vec4 FragColor;

uniform sampler2D texture1;

void main()
{
    // Sample from the texture with proper filtering
    vec4 texColor = texture(texture1, TexCoord);
    
    // For text rendering, we'll use the texture's red channel as the alpha
    // This allows us to render colored text using the font atlas
    // The font texture is grayscale where white (1.0) represents the character
    // Apply proper alpha blending to avoid artifacts
    float alpha = ourColor.a * texColor.r;
    if(alpha < 0.01) discard; // Discard nearly transparent fragments to avoid artifacts
    FragColor = vec4(ourColor.rgb, alpha);
}