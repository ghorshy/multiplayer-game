extends Node

enum State {
	ENTERED,
	CONNECTED,
	INGAME,
	BROWSING_HISCORES,
}

var states_scenes: Dictionary[State, String] = {
	State.ENTERED: "uid://5dlg2fm2yyb2",
	State.CONNECTED: "uid://d05wbkrwbasxi",
	State.INGAME: "uid://c52e2u005l1gn",
	State.BROWSING_HISCORES: "uid://n1bkr4yso6e6"
}

var client_id: int
var current_scene_root: Node

# Game world boundaries received from server
var bounds_min_x: float = -3000.0
var bounds_max_x: float = 3000.0
var bounds_min_y: float = -3000.0
var bounds_max_y: float = 3000.0


func set_state(state: State) -> void:
	if current_scene_root != null:
		current_scene_root.queue_free()
		
	var scene: PackedScene = load(states_scenes[state])
	current_scene_root = scene.instantiate()
	
	add_child(current_scene_root)
