package commands

const Url string = "url"
const Wait string = "wait"
const CaptureImage string = "capture_image"
const SetTarget string = "set_target"
const GetScene string = "get_scene"

// Meshcat-Specific Commands
const SetTransform string = "set_transform"
const SetObject string = "set_object"
const SetProperty string = "set_property"
const Delete string = "delete"
const SetAnimation string = "set_animation"

func IsMeshcatCommand(command string) bool {
	switch command {
	case SetTransform, SetObject, SetProperty, Delete, SetAnimation:
		return true
	default:
		return false
	}
}
