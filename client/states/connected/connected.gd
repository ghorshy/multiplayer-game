extends Node

const packets := preload("uid://gf6c1u38lbn0")

var _action_on_ok_received: Callable

@onready var username_field: LineEdit = $UI/VBoxContainer/UsernameField
@onready var password_field: LineEdit = $UI/VBoxContainer/PasswordField
@onready var login_button: Button = $UI/VBoxContainer/HBoxContainer/LoginButton
@onready var register_button: Button = $UI/VBoxContainer/HBoxContainer/RegisterButton
@onready var _log: Log = $UI/VBoxContainer/Log


func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_received)
	WS.connection_closed.connect(_on_ws_connection_closed)
	login_button.pressed.connect(_on_login_button_pressed)
	register_button.pressed.connect(_on_register_button_pressed)
	

func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_deny_response():
		var deny_response_message := packet.get_deny_response()
		_log.error(deny_response_message.get_reason())
	elif packet.has_ok_response():
		_action_on_ok_received.call()
		

func _on_ws_connection_closed() -> void:
	pass
	

func _on_login_button_pressed() -> void:
	var packet := packets.Packet.new()
	var login_request_message := packet.new_login_request()
	login_request_message.set_username(username_field.text)
	login_request_message.set_password(password_field.text)
	WS.send(packet)
	_action_on_ok_received = func() -> void: GameManager.set_state(GameManager.State.INGAME)
	
	
func _on_register_button_pressed() -> void:
	var packet := packets.Packet.new()
	var register_request_message := packet.new_register_request()
	register_request_message.set_username(username_field.text)
	register_request_message.set_password(password_field.text)
	WS.send(packet)
	_action_on_ok_received = func() -> void: _log.success("Registration successfull")
