extends Area2D

const packets := preload("uid://gf6c1u38lbn0")

const Scene := preload("uid://cifhn2vuxw65m")
const Actor := preload("uid://ddhml87n8qfci")

var actor_id: int
var actor_name: String
var start_x: float
var start_y: float
var start_rad: float
var speed: float
var is_player: bool
var color: Color

var velocity: Vector2
var radius: float:
	set(new_radius):
		radius = new_radius
		collision_shape_2d.set_radius(radius)
		_update_zoom()
		queue_redraw()

var target_zoom := 2.0
var furthest_zoom_allowed := target_zoom
var server_position: Vector2
var time_offset: float = 0.0

@onready var label: Label = $Label
@onready var camera_2d: Camera2D = $Camera2D
@onready var collision_shape_2d: CircleShape2D = $CollisionShape2D.shape

# Shader material for floaty effect
var shader_material: ShaderMaterial

static func instantiate(actor_id: int, actor_name: String, x: float, y: float, radius: float, speed: float, color: Color, is_player: bool) -> Actor:
	var actor := Scene.instantiate()
	actor.actor_id = actor_id
	actor.actor_name = actor_name
	actor.start_x = x
	actor.start_y = y
	actor.start_rad = radius
	actor.speed = speed
	actor.color = color
	actor.is_player = is_player

	return actor


func _input(event: InputEvent) -> void:
	if is_player and event is InputEventMouseButton and event.is_pressed():
		match event.button_index:
			MOUSE_BUTTON_WHEEL_UP:
				target_zoom = min(4, target_zoom + 0.1)
			MOUSE_BUTTON_WHEEL_DOWN:
				target_zoom = max(furthest_zoom_allowed, target_zoom - 0.1)

		camera_2d.zoom.y = camera_2d.zoom.x


func _ready() -> void:
	position.x = start_x
	position.y = start_y
	server_position = position
	velocity = Vector2.RIGHT * speed
	radius = start_rad

	collision_shape_2d.radius = radius
	label.text = actor_name
	time_offset = randf() * TAU  # Random phase offset for each actor

	if not is_player:
		var shader = load("res://objects/actor/floaty.gdshader")
		shader_material = ShaderMaterial.new()
		shader_material.shader = shader
		shader_material.set_shader_parameter("phase_offset", time_offset)
		shader_material.set_shader_parameter("random_seed", randf() * 1000.0)
		material = shader_material


func _process(_delta: float) -> void:
	if not is_equal_approx(camera_2d.zoom.x, target_zoom):
		camera_2d.zoom -= Vector2(1, 1) * (camera_2d.zoom.x - target_zoom) * 0.05


func _physics_process(delta: float) -> void:
	var new_pos := position + velocity * delta

	var buffer := radius
	var rubber_band_zone := 200.0

	var min_x_bound := GameManager.bounds_min_x + buffer
	var max_x_bound := GameManager.bounds_max_x - buffer

	if new_pos.x < min_x_bound:
		new_pos.x = min_x_bound
	elif new_pos.x < min_x_bound + rubber_band_zone:
		var distance_into_boundary := min_x_bound + rubber_band_zone - new_pos.x
		var resistance := distance_into_boundary / rubber_band_zone  # 0 to 1
		var push_back := resistance * resistance * speed * delta * 2.0
		new_pos.x += push_back

	if new_pos.x > max_x_bound:
		new_pos.x = max_x_bound
	elif new_pos.x > max_x_bound - rubber_band_zone:
		var distance_into_boundary := new_pos.x - (max_x_bound - rubber_band_zone)
		var resistance := distance_into_boundary / rubber_band_zone
		var push_back := resistance * resistance * speed * delta * 2.0
		new_pos.x -= push_back

	var min_y_bound := GameManager.bounds_min_y + buffer
	var max_y_bound := GameManager.bounds_max_y - buffer

	if new_pos.y < min_y_bound:
		new_pos.y = min_y_bound
	elif new_pos.y < min_y_bound + rubber_band_zone:
		var distance_into_boundary := min_y_bound + rubber_band_zone - new_pos.y
		var resistance := distance_into_boundary / rubber_band_zone
		var push_back := resistance * resistance * speed * delta * 2.0
		new_pos.y += push_back

	if new_pos.y > max_y_bound:
		new_pos.y = max_y_bound
	elif new_pos.y > max_y_bound - rubber_band_zone:
		var distance_into_boundary := new_pos.y - (max_y_bound - rubber_band_zone)
		var resistance := distance_into_boundary / rubber_band_zone
		var push_back := resistance * resistance * speed * delta * 2.0
		new_pos.y -= push_back

	position = new_pos

	position += (server_position - position) * 0.05

	if not is_player:
		return

	var mouse_pos := get_global_mouse_position()

	var input_vec := position.direction_to(mouse_pos).normalized()
	if abs(velocity.angle_to(input_vec)) > TAU / 30: # 12 degrees
		velocity = input_vec * speed
		var packet := packets.Packet.new()
		var player_direction_message := packet.new_player_direction()
		player_direction_message.set_direction(velocity.angle())
		WS.send(packet)


func _draw() -> void:
	draw_circle(Vector2.ZERO, collision_shape_2d.radius, color)


func _update_zoom() -> void:
	if is_node_ready():
		label.add_theme_font_size_override("font_size", max(16, radius / 2))

	if not is_player:
		return

	var new_furthest_zoom_allowed := 2 * start_rad / radius
	if is_equal_approx(target_zoom, furthest_zoom_allowed):
		target_zoom = new_furthest_zoom_allowed
	furthest_zoom_allowed = new_furthest_zoom_allowed
