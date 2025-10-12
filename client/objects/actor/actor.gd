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

var velocity: Vector2
var radius: float

@onready var label: Label = $Label
@onready var camera_2d: Camera2D = $Camera2D
@onready var collision_shape_2d: CollisionShape2D = $CollisionShape2D

static func instantiate(actor_id: int, actor_name: String, x: float, y: float, radius: float, speed: float, is_player: bool) -> Actor:
	var actor := Scene.instantiate()
	actor.actor_id = actor_id
	actor.actor_name = actor_name
	actor.start_x = x
	actor.start_y = y
	actor.start_rad = radius
	actor.speed = speed
	actor.is_player = is_player
	
	return actor


func _input(event: InputEvent) -> void:
	if is_player and event is InputEventMouseButton and event.is_pressed():
		match event.button_index:
			MOUSE_BUTTON_WHEEL_UP:
				camera_2d.zoom.x = min(4, camera_2d.zoom.x + 0.1)
			MOUSE_BUTTON_WHEEL_DOWN:
				camera_2d.zoom.x = max(0.1, camera_2d.zoom.x - 0.1)
				
		camera_2d.zoom.y = camera_2d.zoom.x


func _ready() -> void:
	position.x = start_x
	position.y = start_y
	velocity = Vector2.RIGHT * speed
	radius = start_rad
	
	collision_shape_2d.radius = radius
	label.text = actor_name
	
	
func _physics_process(delta: float) -> void:
	position += velocity * delta
	
	if not is_player:
		return
		
	# Player-specific stuff
	var mouse_pos := get_global_mouse_position()
	
	var input_vec := position.direction_to(mouse_pos).normalized()
	if abs(velocity.angle_to(input_vec)) > TAU / 15: # 24 degrees
		velocity = input_vec * speed
		var packet := packets.Packet.new()
		var player_direction_message := packet.new_player_direction()
		player_direction_message.set_direction(velocity.angle())
		WS.send(packet)


func _draw() -> void:
	draw_circle(Vector2.ZERO, collision_shape_2d.radius, Color.DARK_ORCHID)
