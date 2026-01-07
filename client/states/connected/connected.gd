extends Node

const packets := preload("uid://gf6c1u38lbn0")

var _action_on_ok_received: Callable

@onready var _log: Log = %Log
@onready var login_form: LoginForm = %LoginForm
@onready var register_form: RegisterForm = %RegisterForm
@onready var register_prompt: RichTextLabel = %RegisterPrompt


func _ready() -> void:
	WS.packet_received.connect(_on_ws_packet_received)
	WS.connection_closed.connect(_on_ws_connection_closed)
	login_form.form_submitted.connect(_on_login_form_submitted)
	register_form.form_submitted.connect(_on_register_form_submitted)
	register_form.form_cancelled.connect(_on_register_form_cancelled)
	register_prompt.meta_clicked.connect(_on_register_prompt_meta_clicked)
	

func _on_ws_packet_received(packet: packets.Packet) -> void:
	var sender_id := packet.get_sender_id()
	if packet.has_deny_response():
		var deny_response_message := packet.get_deny_response()
		_log.error(deny_response_message.get_reason())
	elif packet.has_ok_response():
		_action_on_ok_received.call()
		

func _on_ws_connection_closed() -> void:
	pass
	

func _on_login_form_submitted(username: String, password: String) -> void:
	if password.is_empty():
		_log.error("Password cannot be empty")
		return

	var packet := packets.Packet.new()
	var login_request_msg := packet.new_login_request()
	login_request_msg.set_username(username)
	login_request_msg.set_password(password)
	WS.send(packet)
	_action_on_ok_received = func() -> void: GameManager.set_state(GameManager.State.INGAME)


func _on_register_form_submitted(username: String, password: String, confirm_password: String, color: Color) -> void:
	if password.is_empty():
		_log.error("Password cannot be empty")
		return

	if password != confirm_password:
		_log.error("Passwords do not match")
		return

	var packet := packets.Packet.new()
	var register_request_msg := packet.new_register_request()
	register_request_msg.set_username(username)
	register_request_msg.set_password(password)
	register_request_msg.set_color(color.to_rgba32())
	WS.send(packet)
	_action_on_ok_received = func() -> void:
		_log.success("Registration successful! You can now log in.")
		register_form.hide()
		login_form.show()
		register_prompt.show()


func _on_register_form_cancelled() -> void:
	register_form.hide()
	login_form.show()
	register_prompt.show()


func _on_register_prompt_meta_clicked(meta: Variant) -> void:
	if meta is String and meta == "register":
		login_form.hide()
		register_form.show()
		register_prompt.hide()
