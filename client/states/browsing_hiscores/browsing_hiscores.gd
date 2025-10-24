extends Node

const packets := preload("res://packets.gd")

@onready var back_button: Button = %BackButton
@onready var line_edit: LineEdit = %LineEdit
@onready var search_button: Button = %SearchButton
@onready var hiscores: Hiscores = %Hiscores
@onready var _log: Log = %Log


func _ready() -> void:
	line_edit.text_submitted.connect(_on_line_edit_text_submitted)
	search_button.pressed.connect(_on_search_button_pressed)
	back_button.pressed.connect(_on_back_button_presesed)
	WS.packet_received.connect(_on_ws_packet_received)
	
	var packet := packets.Packet.new()
	packet.new_hi_score_board_request()
	WS.send(packet)
	
func _on_ws_packet_received(packet: packets.Packet) -> void:
	if packet.has_hiscore_board():
		_handle_hiscore_board_msg(packet.get_hiscore_board())
		
		
func _handle_hiscore_board_msg(hiscore_board_msg: packets.HiscoreBoardMessage) -> void:
	hiscores.clear_hiscores()
	for hiscore_msg: packets.HiscoreMessage in hiscore_board_msg.get_hiscores():
		var name := hiscore_msg.get_name()
		var rank_and_name := "%d. %s" % [hiscore_msg.get_rank(), name]
		var score := hiscore_msg.get_score()
		var highlight := name.to_lower() == line_edit.text.to_lower()
		hiscores.set_hiscore(rank_and_name, score, highlight)
		
		
func _handle_deny_response(deny_response_msg: packets.DenyResponseMessage) -> void:
	_log.error(deny_response_msg.get_reason())


func _on_line_edit_text_submitted(new_text: String) -> void:
	_on_search_button_pressed()


func _on_search_button_pressed() -> void:
	var packet := packets.Packet.new()
	var search_hiscore_msg := packet.new_search_hiscore()
	search_hiscore_msg.set_name(line_edit.text)
	WS.send(packet)


func _on_back_button_presesed() -> void:
	var packet := packets.Packet.new()
	packet.new_finished_browsing_hiscores()
	WS.send(packet)
	GameManager.set_state(GameManager.State.CONNECTED)
