package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"

	pb "Client/proto"
)

func main() {
	// Read configuration from environment variables.
	controlServerAddr := os.Getenv("CONTROL_SERVER")
	if controlServerAddr == "" {
		controlServerAddr = "localhost:50051"
	}
	robotIDStr := os.Getenv("ROBOT_ID")
	if robotIDStr == "" {
		log.Fatal("ROBOT_ID not set")
	}
	robotID, err := strconv.Atoi(robotIDStr)
	if err != nil {
		log.Fatalf("Invalid ROBOT_ID: %v", err)
	}

	// Connect to the control server.
	conn, err := grpc.Dial(controlServerAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to control server: %v", err)
	}
	defer conn.Close()
	client := pb.NewControlServiceClient(conn)
	log.Printf("Connected to control server at %s", controlServerAddr)

	// Start sending streams (client-streaming RPCs).
	go sendRobotStatus(client, int32(robotID))
	go sendLocationStatus(client, int32(robotID))
	go sendWaypointStream(client, int32(robotID))

	// Start receiving streams (server-streaming RPCs).
	go receiveRobotSetModeStream(client, int32(robotID))
	go receiveWaypointsStream(client, int32(robotID))
	go receivePeripheralControlStream(client, int32(robotID))
	go receiveJoystickControlStream(client, int32(robotID))

	// Block forever.
	select {}
}

// sendRobotStatus opens a client-streaming RPC to send RobotStatusData.
func sendRobotStatus(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	stream, err := client.SendRobotStatus(ctx)
	if err != nil {
		log.Printf("Error opening SendRobotStatus stream: %v", err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		msg := &pb.RobotStatusData{
			RobotId:  robotID,
			ArmState: true,
			VehicleControlMode: &pb.VehicleControlMode{
				Manual:   true,
				Offboard: false,
				Auto:     false,
			},
			ClearbotControlMode: &pb.ClearbotControlMode{
				JoystickControl:   true,  // Updated to CamelCase
				ObstacleAvoidance: false, // Updated to CamelCase
				HeadingControl:    false, // Updated to CamelCase
				WaypointControl:   true,  // Updated to CamelCase
			},
			Peripherals: &pb.Peripherals{
				Conveyor: true,
			},
			SystemInfo: &pb.SystemInfo{
				Temperature: 36.5,
			},
		}
		if err := stream.Send(msg); err != nil {
			log.Printf("Error sending RobotStatusData: %v", err)
			return
		}
		log.Printf("Sent RobotStatusData: %+v", msg)
	}
	// Optionally, call stream.CloseAndRecv() when done.
}

// sendLocationStatus opens a client-streaming RPC to send LocationStatusData.
func sendLocationStatus(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	stream, err := client.SendLocationStatus(ctx)
	if err != nil {
		log.Printf("Error opening SendLocationStatus stream: %v", err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		msg := &pb.LocationStatusData{
			RobotId: robotID,
			CurrentPosition: &pb.Coordinates{
				Latitude:  12.34,
				Longitude: 56.78,
				Altitude:  100.0,
			},
			HomePosition: &pb.Coordinates{
				Latitude:  12.30,
				Longitude: 56.70,
				Altitude:  95.0,
			},
			Velocity: &pb.Vector3F{
				X: 1.0,
				Y: 0.0,
				Z: 0.0,
			},
			Heading: 90.0,
		}
		if err := stream.Send(msg); err != nil {
			log.Printf("Error sending LocationStatusData: %v", err)
			return
		}
		log.Printf("Sent LocationStatusData: %+v", msg)
	}
}

// sendWaypointStream opens a client-streaming RPC to send WaypointStreamData.
func sendWaypointStream(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	stream, err := client.SendWaypointStream(ctx)
	if err != nil {
		log.Printf("Error opening SendWaypointStream stream: %v", err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		msg := &pb.WaypointStreamData{
			RobotId: robotID,
			Waypoints: &pb.ClearbotWaypoints{
				CurrentWp: &pb.Waypoint{
					X: 10.0,
					Y: 20.0,
				},
				RemainingWps: []*pb.Waypoint{
					{X: 11.0, Y: 21.0},
					{X: 12.0, Y: 22.0},
				},
				FinishedWps: []*pb.Waypoint{
					{X: 9.0, Y: 19.0},
				},
			},
			StartPosition: &pb.Coordinates{
				Latitude:  12.34,
				Longitude: 56.78,
				Altitude:  100.0,
			},
			RefPosition: &pb.Coordinates{
				Latitude:  12.30,
				Longitude: 56.70,
				Altitude:  95.0,
			},
			WpDist: 5.5,
		}
		if err := stream.Send(msg); err != nil {
			log.Printf("Error sending WaypointStreamData: %v", err)
			return
		}
		log.Printf("Sent WaypointStreamData: %+v", msg)
	}
}

// receiveRobotSetModeStream uses a server-streaming RPC to receive GetRobotSetModeData.
func receiveRobotSetModeStream(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	req := &pb.GetRobotSetModeRequest{RobotId: robotID}
	stream, err := client.GetRobotSetModeStream(ctx, req)
	if err != nil {
		log.Printf("Error calling GetRobotSetModeStream: %v", err)
		return
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving GetRobotSetModeData: %v", err)
			return
		}
		log.Printf("Received GetRobotSetModeData: mode=%s", msg.Mode)
	}
}

// receiveWaypointsStream uses a server-streaming RPC to receive WaypointsData.
func receiveWaypointsStream(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	req := &pb.GetWaypointsRequest{RobotId: robotID}
	stream, err := client.GetWaypointsStream(ctx, req)
	if err != nil {
		log.Printf("Error calling GetWaypointsStream: %v", err)
		return
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving WaypointsData: %v", err)
			return
		}
		log.Printf("Received WaypointsData: %d waypoints", len(msg.Waypoints))
	}
}

// receivePeripheralControlStream uses a server-streaming RPC to receive Peripherals.
func receivePeripheralControlStream(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	req := &pb.GetPeripheralControlRequest{RobotId: robotID}
	stream, err := client.GetPeripheralControlStream(ctx, req)
	if err != nil {
		log.Printf("Error calling GetPeripheralControlStream: %v", err)
		return
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving Peripherals: %v", err)
			return
		}
		log.Printf("Received Peripherals: conveyor=%v", msg.Conveyor)
	}
}

// receiveJoystickControlStream uses a server-streaming RPC to receive VehicleThrustTorque.
func receiveJoystickControlStream(client pb.ControlServiceClient, robotID int32) {
	ctx := context.Background()
	req := &pb.GetJoystickControlRequest{RobotId: robotID}
	stream, err := client.GetJoystickControlStream(ctx, req)
	if err != nil {
		log.Printf("Error calling GetJoystickControlStream: %v", err)
		return
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Error receiving VehicleThrustTorque: %v", err)
			return
		}
		log.Printf("Received VehicleThrustTorque: throttle=%.2f, yaw=%.2f", msg.Throttle, msg.Yaw)
	}
}
