shader_type canvas_item;

uniform float wobble_speed = 1.5;
uniform float wobble_amount = 0.05;
uniform float phase_offset = 0.0;
uniform float random_seed = 0.0;

// Hash function for pseudo-random values
float hash(float n) {
    return fract(sin(n) * 43758.5453123);
}

void vertex() {
    // Agar.io style wobble: random floaty deformation
    vec2 center = vec2(0.5, 0.5);
    vec2 uv_centered = UV - center;

    // Calculate angle from center
    float angle = atan(uv_centered.y, uv_centered.x);

    // Create pseudo-random offsets for this specific object using random_seed
    float r1 = hash(random_seed);
    float r2 = hash(random_seed + 1.0);
    float r3 = hash(random_seed + 2.0);
    float r4 = hash(random_seed + 3.0);
    float r5 = hash(random_seed + 4.0);
    float r6 = hash(random_seed + 5.0);

    // Multiple sine waves with random-ish parameters for organic wobble
    // Each wave has a different frequency, phase, and direction
    float wobble = 0.0;
    wobble += sin(TIME * wobble_speed * (0.8 + r1 * 0.4) + angle * (2.0 + r2 * 3.0) + phase_offset) * wobble_amount;
    wobble += sin(TIME * wobble_speed * (1.2 + r3 * 0.6) - angle * (3.0 + r4 * 2.0) + phase_offset * 1.7) * wobble_amount * 0.7;
    wobble += sin(TIME * wobble_speed * (0.5 + r5 * 0.8) + angle * (4.0 + r6 * 3.0) + phase_offset * 0.5) * wobble_amount * 0.5;
    wobble += sin(TIME * wobble_speed * (1.8 + r1 * 0.5) - angle * (1.5 + r2 * 2.5) + phase_offset * 2.3) * wobble_amount * 0.4;

    // Apply wobble to the vertex position by scaling from center
    float distance_from_center = length(uv_centered);
    if (distance_from_center > 0.001) {
        vec2 direction = normalize(uv_centered);
        VERTEX += direction * wobble * 100.0;
    }
}

void fragment() {
    COLOR = texture(TEXTURE, UV) * COLOR;
}
