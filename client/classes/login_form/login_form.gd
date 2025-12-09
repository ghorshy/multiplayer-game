class_name LoginForm
extends VBoxContainer


signal form_submitted(username: String, password: String)


@onready var username_field: LineEdit = %UsernameField
@onready var password_field: LineEdit = %PasswordField
@onready var login_button: Button = %LoginButton
@onready var hiscores_button: Button = %HiscoresButton


func _ready() -> void:
	login_button.pressed.connect(_on_login_button_pressed)
	hiscores_button.pressed.connect(_on_hiscores_button_pressed)
	username_field.text_submitted.connect(_on_field_submitted)
	password_field.text_submitted.connect(_on_field_submitted)
	
	
func _on_login_button_pressed() -> void:
	form_submitted.emit(username_field.text, password_field.text)


func _on_hiscores_button_pressed() -> void:
	GameManager.set_state(GameManager.State.BROWSING_HISCORES)


func _on_field_submitted(_text: String) -> void:
	_on_login_button_pressed()
