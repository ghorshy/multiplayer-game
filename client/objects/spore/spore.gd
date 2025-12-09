extends Area2D


const Scene := preload("uid://btoc3785fenoe")
const Spore := preload("uid://dichmp4fpj04i")
const Actor := preload("uid://ddhml87n8qfci")

var spore_id: int
var x: float
var y: float
var radius: float
var color: Color
var underneath_player: bool
var time_offset: float = 0.0

@onready var collision_shape_2d: CircleShape2D = $CollisionShape2D.shape

# Shader material for floaty wobble effect
var shader_material: ShaderMaterial


static func instantiate(spore_id: int, x: float, y: float, radius: float, underneath_player: bool) -> Spore:
	var spore := Scene.instantiate() as Spore
	spore.spore_id = spore_id
	spore.x = x
	spore.y = y
	spore.radius = radius
	spore.underneath_player = underneath_player
	
	return spore
	
	
func _ready() -> void:
	if underneath_player:
		area_exited.connect(_on_area_exited)
	position.x = x
	position.y = y
	collision_shape_2d.radius = radius

	color = Color.from_hsv(randf(), 1, 1, 1)
	time_offset = randf() * TAU  # Random phase offset for each spore

	# Load and setup floaty wobble shader
	var shader = load("res://objects/actor/floaty.gdshader")
	shader_material = ShaderMaterial.new()
	shader_material.shader = shader
	shader_material.set_shader_parameter("phase_offset", time_offset)
	shader_material.set_shader_parameter("random_seed", randf() * 1000.0)
	shader_material.set_shader_parameter("wobble_amount", 0.06)  # Slightly more wobble for spores
	material = shader_material
	

func _draw() -> void:
	draw_circle(Vector2.ZERO, radius, color)
	
	
func _on_area_exited(area: Area2D) -> void:
	if area is Actor:
		underneath_player = false
