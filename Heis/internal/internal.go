package internal

import (
	. ".././driver"
	. "fmt"
	"time"
)

var speed int

func open_doors() {
	Heis_set_door_open_lamp(1)
	time.Sleep(1000 * time.Millisecond)
	Heis_set_door_open_lamp(0)
	return
}
func Init_orders() [][]int {
	orders := make([][]int, 5)
	for i := 0; i < 5; i++ {
		orders[i] = make([]int, 4)
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 4; j++ {
			orders[i][j] = 0
		}
	}
	Print("Orders:")
	Println("")
	Print_orders(orders)
	return orders
}

func Print_orders(orders [][]int) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 3; j++ {
			Print(orders[i][j])
			Print("     ")
		}
		println("")
	}
	println("")
}
func remove_from_queue(queue []int) []int {

	for i := 0; i < 3; i++ {
		queue[i] = queue[i+1]
	}
	queue[3] = -1
	return queue
}
func remove_from_orders(orders [][]int, current_order int) [][]int {
	if current_order != -1 {
		for i := 0; i < 3; i++ {
			orders[current_order][i] = 0
		}
	}
	set_external_lights(orders)
	set_internal_lights(orders)
	return orders
}
func Send_to_floor(queue []int, orders [][]int) ([]int, [][]int) {
	for i := 0; i < 4; i++ {
		current_order := queue[0]
		current_floor := Heis_get_floor()
		if current_order == -1 {
			Heis_set_speed(0)
		}
		if current_floor == -1 {
			stop_all()
		}
		if Heis_get_floor() != -1 && current_order != -1 {
			Print("Current floor: ")
			Println(current_floor + 1)
			Print("Going to: Floor ")
			Println(current_order + 1)
		}
		if current_order == current_floor && current_order != -1 {
			open_doors()
			queue = remove_from_queue(queue)
			orders = remove_from_orders(orders, current_order)
			current_order = -1
		}
		if current_order > current_floor && current_order != -1 {
			Heis_set_speed(speed)
			for {
				if Heis_get_floor() == current_order {
					Heis_set_speed(0)
					Print("Arrived at floor: ")
					Println(current_order + 1)
					open_doors()
					queue = remove_from_queue(queue)
					orders = remove_from_orders(orders, current_order)
					return queue, orders
				}
			}
		}
		if current_order < current_floor && current_order != -1 {
			Heis_set_speed(-speed)
			for {
				if Heis_get_floor() == current_order {
					Heis_set_speed(0)
					Print("Arrived at floor: ")
					Println(current_order + 1)
					open_doors()
					queue = remove_from_queue(queue)
					orders = remove_from_orders(orders, current_order)
					return queue, orders
				}
			}
		}
	}
	return queue, orders
}
func get_queue(orders [][]int) []int {
	queue := make([]int, 4)
	for i := 0; i < 4; i++ {
		queue[i] = -1
	}
	k := 0
	for i := 0; i < 4; i++ {

		if orders[i][0] == 1 || orders[i][1] == 1 || orders[i][2] == 1 {
			queue[k] = i
			Println("Order received: Floor ")
			Println(i + 1)
			k++
		}

	}
	return queue
}

//Handles external button presses
func get_orders(orders [][]int) {
	for {
		for floor := 0; floor < 4; floor++ {

			if floor != 3 {
				if Heis_get_button(BUTTON_CALL_UP, floor) == 1 {
					orders[floor][BUTTON_CALL_UP] = 1

				}
			}
			if Heis_get_button(BUTTON_COMMAND, floor) == 1 {
				orders[floor][BUTTON_COMMAND] = 1
			}
			if floor != 0 {
				if Heis_get_button(BUTTON_CALL_DOWN, floor) == 1 {
					orders[floor][BUTTON_CALL_DOWN] = 1

				}
			}
		}
		set_internal_lights(orders)
		set_external_lights(orders)
	}
}

//Checks which floor the elevator is on and sets the floor-light
func Floor_indicator() {
	Println("executing floor indicator!")
	var floor int
	for {
		floor = Heis_get_floor()
		//Println(floor)
		if floor != -1 {
			Heis_set_floor_indicator(floor)
			//time.Sleep(50 * time.Millisecond)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
func To_nearest_floor() {
	for {

		Heis_set_speed(-speed)
		if Heis_get_floor() == 0 {
			Heis_set_speed(0)

			return
		}
		if Heis_get_floor() != -1 {
			Heis_set_speed(0)
			return
		}
	}
}
func get_stop() {
	for {
		if Heis_stop() {
			Heis_set_stop_lamp(1)
			stop_all()
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func stop_all() {
	Heis_set_speed(0)
	time.Sleep(time.Millisecond * 1000)
	Heis_set_stop_lamp(0)
	To_nearest_floor()
	Heis_init()
}

func set_internal_lights(orders [][]int) {
	for i := 0; i < 4; i++ {
		if i != 3 {
			Heis_set_button_lamp(BUTTON_CALL_UP, i, orders[i][0])
		}
		if i != 0 {
			Heis_set_button_lamp(BUTTON_CALL_DOWN, i, orders[i][1])
		}
	}
}
func set_external_lights(orders [][]int) {
	for i := 0; i < 4; i++ {
		Heis_set_button_lamp(BUTTON_COMMAND, i, orders[i][2])
	}
}
func Internal() {
	// Init
	speed = 150
	orders := Init_orders()
	queue := get_queue(orders)
	Heis_init()
	Heis_set_speed(0)
	To_nearest_floor()
	Heis_set_stop_lamp(0)

	go get_stop()
	go Floor_indicator()
	go get_orders(orders)
	for {
		queue = get_queue(orders)
		queue, orders = Send_to_floor(queue, orders)
	}

	select {}
}
