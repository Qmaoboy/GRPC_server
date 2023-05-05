package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "Voting"

	"github.com/jamesruan/sodium"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

// voting service

func create_Election(ctx context.Context, client pb.EVotingClient) {
	var e_name string
	var e_time int
	var group []string
	var choice []string
	fmt.Println("Start to Creat Election")
	fmt.Printf("Please input the Title: ")
	fmt.Scan(&e_name)
	fmt.Printf("How long is the election exitst ?: ")
	fmt.Scan(&e_time)
	tt := time.Now().Add(time.Minute * time.Duration(e_time))
	end_time := timestamppb.New(tt)
	t1 := time.Unix(end_time.GetSeconds(), 0)
	fmt.Printf("end Time: ")
	fmt.Println(t1)
	fmt.Println("Please input group which involve(e for end): ")
	for {
		var temp string
		fmt.Scan(&temp)
		switch temp {
		case "e":
			break
		default:
			in := true
			for _, e := range group {
				if e == temp {
					in = false
					break
				}
			}
			if in {
				group = append(choice, temp)
			}
		}
	}
	fmt.Println("Please input Choice which involve(e for end): ")
	for {
		var temp string
		fmt.Scan(&temp)
		switch temp {
		case "e":
			break
		default:
			fmt.Println("")
			in := true
			for _, e := range choice {
				if e == temp {
					in = false
					break
				}
			}
			if in {
				choice = append(choice, temp)
			}
		}
	}
	// e := Celection_info{
	// 	Name:     &e_name,
	// 	Choices:  choice,
	// 	end_time: end_time,
	// }
	_, err := client.CreateElection(ctx, &pb.Election{
		Name:    &e_name,
		Groups:  group,
		Choices: choice,
		EndDate: end_time,
		Token:   &pb.AuthToken{value: Vpublic_key},
	})
	if err != nil {
		fmt.Printf("Creat Election error")
		fmt.Println(err)
	}
}

func cast_Vote(ctx context.Context, client pb.EVotingClient) {
	var e_name, e_choice string
	fmt.Printf("Please input the Election Name ?: ")
	fmt.Scan(&e_name)
	for _, v := range Elect_list {
		now := time.Now()
		if e_name == v.Name && v.end_time.Before(now) {
			for _, c := range v.Choices {
				fmt.Println(c)
			}
			break
		} else {
			fmt.Println("No Such Alive Election!!")
			return
		}
	}
	fmt.Printf("Please input the Choice you want to Vote?")
	fmt.Scan(&e_choice)
	s, err := client.CastVote(ctx, &pb.Vote{
		Election_name: e_name,
		Choice_name:   e_choice,
		Token:         &pb.Authtoken{value: Vpublic_key},
	})
	fmt.Printf("Result: %s", s)
	if err != nil {
		fmt.Printf("Cast Vote error")
		fmt.Println(err)
	}
}
func get_result(ctx context.Context, client pb.EVotingClient) {
	var e_name string
	fmt.Printf("What Election Do you want to see ?: ")
	fmt.Scan(&e_name)
	r, err := client.Get_result(ctx, &pb.ElectionName{name: &e_name})
	if err != nil {
		fmt.Printf("Get Result Error error")
		fmt.Println(err)
	} else {
		fmt.Printf("Election Result")
		fmt.Println(r)
	}
}
func stdinput(s string) string {
	// var buffet [512]byte
	stdintext := ""
	fmt.Printf("%s", s)
	Reader := bufio.NewReader(os.Stdin)
	stdintext, _ = Reader.ReadString('\n')
	// fmt.Printf(stdintext)
	return stdintext
}

type RVoter_info struct {
	ip          string
	port        string
	Name        string
	Group       string
	Public_key  sodium.SignPublicKey
	Private_key sodium.SignSecretKey
	V_token     []byte
}

func main() {
	var voter RVoter_info
	fmt.Printf("Please Server IP: ")
	fmt.Scan(&voter.ip)
	fmt.Printf("Please Server Port: ")
	fmt.Scan(&voter.port)
	fmt.Printf("what's your name: ")
	fmt.Scan(&voter.Name)
	fmt.Printf("what's your group: ")
	fmt.Scan(&voter.Group)

	fmt.Println("Client Starting to do the service....")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
	defer cancel()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", voter.ip, voter.port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer conn.Close()
	c1 := pb.NewEVotingClient(conn)
	if err != nil {
		log.Fatalln(err)
	} else {
		log.Printf("Wellcome to the Evoting Server!! Connect Sceuss fully")
	}
	// tokenByte=Auth(ctx,c1)
	var in string
	for {
		fmt.Println("Please input the task tou want to perform(e:exit)")
		time.Sleep(1 * time.Nanosecond)
		listall()
		fmt.Printf("Your answer: ")
		fmt.Scan(&in)
		switch in {
		case "f":
			voter = Change_Voter_info()
		case "e":
			fmt.Println("Exit the client")
			os.Exit(0)
		case "c":
			create_Election(ctx, c1)
		case "v":
			cast_Vote(ctx, c1)
		case "a":
			voter.V_token = Auth(ctx, c1, voter)
		case "g":
			get_result(ctx, c1)
		default:
			fmt.Println("Unknown Task")
		}

	}

}

func listall() {

	fmt.Println("e:exit")
	fmt.Println("f:Change Voter info")
	fmt.Println("v:Cast Vote")
	fmt.Println("g:get_result")
	fmt.Println("c:Creat Election")

}
func Change_Voter_info() RVoter_info {
	var voter RVoter_info
	fmt.Printf("what's your name: ")
	fmt.Scan(&voter.Name)
	fmt.Printf("what's your group: ")
	fmt.Scan(&voter.Group)
	return voter
}

func Auth(ctx context.Context, client pb.EVotingClient, vv RVoter_info) (sodium.SignPublicKey, sodium.SignSecretKey, []byte) {
	fmt.Println("Start to Authentication")

	votername := vv.Name
	cc, err := client.PreAuth(ctx, &pb.VoterName{Name: &votername})
	if err != nil {
		log.Printf("PreAuth Fail")
		log.Println(err)
		os.Exit(1)
	}
	challenge := cc.Value
	/////////////////////////////// 加密
	seed := sodium.SignSeed{
		Bytes: challenge,
	}
	key := sodium.SeedSignKP(seed)
	///////////////////////////////
	response := pb.AuthRequest{
		Name:     &pb.VoterName{Name: &votername},
		Response: nil,
	}
	Authtoken, err := client.Auth(ctx, &response)
	if err != nil {
		fmt.Printf("Authorize Fail")
		fmt.Println(err)
		os.Exit(1)
	}
	return key.Publickey, key.SecretKey, Authtoken.Value

}

func pre_auth(ctx context.Context, client pb.EVotingClient)
