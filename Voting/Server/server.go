package main

import (
	pb "Voting"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/jamesruan/sodium"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Voter_info struct {
	Name           string
	Group          string
	Public_key     sodium.SignPublicKey
	Private_key    sodium.SignSecretKey
	V_token        []byte // rand
	Key_seed       []byte // rand
	Challenge      []byte // rand
	Alive          bool
	token_End_time time.Time
}

type election_info struct {
	Name          string
	Groups        []string
	Choices       []string
	Already_voted []string
	End_time      time.Time
	Alive         bool
	Result        map[string]int32
}

var (
	voter_list    []Voter_info
	election_list []election_info
)

// Function
func RegisterVoter() (int, error) {
	var name, group string
	fmt.Println("Start to  register Voter")
	fmt.Println("Input Votername: ")
	fmt.Scan(&name)
	fmt.Println("Input Votername: ")
	fmt.Scan(&group)

	VID := make([]byte, 10)
	_, err := rand.Read(VID)
	if err != nil {
		// 处理错误
	}

	////////////////////////////////////////////////////// Creat Key Pair
	seed := sodium.SignSeed{Bytes: VID}
	key := sodium.SeedSignKP(seed)
	//////////////////////////////////////////////////////
	v_info := Voter_info{
		Name:        name,
		Group:       group,
		Public_key:  key.PublicKey,
		Private_key: key.SecretKey,
		Key_seed:    VID,
		Alive:       true,
	}
	change := true
	for _, v := range voter_list {
		if v.Name == name {
			v.Alive = true
			voter_list = append(voter_list, v_info)
			change = false
		}
	}

	if change {
		return 0, nil
	} else {
		return 1, nil
	}
}

func UnregisterVoter() (int, error) {
	var name string
	fmt.Println("Stating to unregister Voter")
	fmt.Printf("Please input Voter' name: ")
	fmt.Scanln(&name)
	change := true
	for i, v_ := range voter_list {
		if v_.Name == name {
			voter_list = append(voter_list[:i], voter_list[i+1:]...)
			change = false
		}
	}
	if change {
		return 1, nil
	} else {
		return 0, nil
	}

}

type Voting struct {
	pb.UnimplementedEVotingServer
}

func listall() {
	fmt.Println("r : register voter")
	fmt.Println("d : unregister voter")
	fmt.Println("w : Show_voter")
	fmt.Println("i : Show_election")
	fmt.Println("f : Finish register")
	fmt.Println("e : exit")
}

// listen
func main() {
	var ip, port string

	fmt.Printf("please input IP: ")
	fmt.Scanln(&ip)
	fmt.Printf("please input port: ")
	fmt.Scanln(&port)
	fmt.Printf("starting gRPC server... at %s....\n", ip)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", ip, port))

	if err != nil {
		log.Fatalf("Fail to listen: %v", err)
	}

	log.Println("GRPC server is running")
	log.Printf("Server is Listning at %v", lis.Addr())

	grpcserver := grpc.NewServer()
	pb.RegisterEVotingServer(grpcserver, &Voting{})
	reflection.Register(grpcserver)

	go func() {
		if err := grpcserver.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	for {
		var in string
		fmt.Printf("What task do you want to perform (l:list / e:exit)?: ")
		fmt.Scan(&in)
		b := false
		switch in {
		case "r":
			RegisterVoter()
		case "l":
			listall()
		case "d":
			UnregisterVoter()
		case "e":
			fmt.Println("exit the server")
			os.Exit(0)
		case "f":
			fmt.Println("complete registration")
			b = true
		case "w":
			Who_voter()
		case "i":
			Who_voter()
		default:
			fmt.Println("No Function !!")
		}
		time.Sleep(1 * time.Nanosecond)
		if b {
			break
		}
	}
}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func authorize_token(token []byte) int32 {
	now := time.Now()
	if token == nil {
		return 1
	}
	for _, vo := range voter_list {
		if Equal(token, vo.V_token) {
			if vo.token_End_time.Before(now) {
				return 1
			}
		}
	}
	return 0

}
func Check_election_Time(electioname string) bool {
	now := time.Now()
	for _, vo := range election_list {
		if vo.Name == electioname {
			if vo.End_time.Before(now) {
				return false
			}
		}
	}
	return true
}

func (s *Voting) CreateElection(ctx context.Context, in *pb.Election) (*pb.Status, error) {
	log.Println("Start to Creat Election")
	///////////////////////////////////////////////Check token
	token := in.Token.Value
	auth := authorize_token(token)
	if auth == 1 {
		fmt.Println("Token Missing")
		return &pb.Status{Code: &auth}, nil
	} else if auth == 1 {
		fmt.Println("Token Expired!")
		return &pb.Status{Code: &auth}, nil
	} else {
		log.Println("Auth Success!!")
	}
	//////////////////////////////////////////////// Find Group and choices
	var co = int32(0)
	if in.Groups == nil || in.Choices == nil {
		fmt.Println("Group or Choices missing")
		co = 2
		return &pb.Status{Code: &co}, nil
	}

	//////////////////////////////////////////////// init election para
	end_time := in.EndDate
	t1 := time.Unix(end_time.GetSeconds(), 0)

	result := map[string]int32{}
	for _, ch := range in.Choices {
		result[ch] = int32(0)
	}
	////////////////////////////////////////////////
	create_e := election_info{
		Name:     *in.Name,
		Groups:   in.Groups,
		Choices:  in.Choices,
		End_time: t1,
		Alive:    true,
		Result:   result,
	}
	election_list = append(election_list, create_e)
	co = 0
	fmt.Println("Creat Success")
	who_elect()
	return &pb.Status{Code: &co}, nil
}

func (s *Voting) CastVote(ctx context.Context, in *pb.Vote) (*pb.Status, error) {
	co := int32(0)
	//////////////////////////////////////////////////////////////////
	token := in.Token.Value
	auth := authorize_token(token)
	if auth == 1 {
		fmt.Println("Token Missing")
		return &pb.Status{Code: &auth}, nil
	} else if auth == 1 {
		fmt.Println("Token Expired!")
		return &pb.Status{Code: &auth}, nil
	} else {
		log.Println("Auth Success!!")
	}
	////////////////////////////////////////////////////////////////// Authorized
	if !Check_election_Time(*in.ElectionName) {
		log.Println("Election Expired !!!")
		co = 1
		return &pb.Status{Code: &co}, nil
	}
	//////////////////////////////////////////////////////////////////
	vote := false
	for _, vo := range election_list {
		if vo.Name == *in.ElectionName {
			for _, cc := range vo.Choices {
				if cc == *in.ChoiceName {
					vo.Result[cc] += 1
					vote = true
				}
			}
		}
	}
	if !vote {
		co = 2
		return &pb.Status{Code: &co}, nil
	}
	return &pb.Status{Code: &co}, nil
}

func (s *Voting) GetResult(ctx context.Context, in *pb.ElectionName) (*pb.ElectionResult, error) {
	election_name := *in.Name
	var counts []*pb.VoteCount
	success := false
	var co int32
	for _, ele := range election_list {
		if election_name == ele.Name {
			now := time.Now()
			if ele.End_time.Before(now) {
				success = true
				for _, ch := range ele.Choices {
					fmt.Print("choice: ")
					fmt.Println(ch)
					t := ele.Result[ch]
					c_name := ch
					temp := pb.VoteCount{ChoiceName: &c_name, Count: &t}
					counts = append(counts, &temp)
				}
				fmt.Println(counts)
			} else {
				fmt.Println("time is not up yet")
				co = 2
				counts = nil
				return &pb.ElectionResult{Status: &co, Counts: counts}, nil
			}
		}
	}

	co = 0
	if !success {
		fmt.Println("Non-existent election")
		co = 1
		return &pb.ElectionResult{Status: &co, Counts: counts}, nil
	}
	return &pb.ElectionResult{Status: &co, Counts: counts}, nil
}

func Who_voter() {
	if len(voter_list) == 0 {
		fmt.Println("No Voter alive")
	}
	for _, v := range voter_list {
		fmt.Println(v.Name)
	}
}
func who_elect() {
	if len(election_list) == 0 {
		fmt.Println("No Election alive")
	}
	for _, e := range election_list {
		fmt.Println(e)
	}
}

func (c *Voting) Auth(ctx context.Context, in *pb.AuthRequest) (*pb.AuthToken, error) {
	name := *in.Name.Name
	V_token := make([]byte, 10)
	_, Verr := rand.Read(V_token)
	if Verr != nil {
		fmt.Println("Token error")
	}

	sig := sodium.Signature{
		Bytes: in.Response.Value,
	}
	find := false
	for _, vo := range voter_list {
		if vo.Name == name {
			find = true
			ch := sodium.Bytes(vo.Challenge)
			err := ch.SignVerifyDetached(sig, vo.Public_key)
			if err != nil {
				log.Println("Auth error")
				return nil, nil
			} else {
				log.Println("The challenge is properly signed")
				now := time.Now()
				end_time := now.Add(time.Hour * 1)
				fmt.Print("token valid time: ")
				fmt.Println(end_time)
				vo.token_End_time = end_time
				vo.V_token = V_token
				break
			}
		}
	}
	if find {
		return &pb.AuthToken{Value: V_token}, nil
	}
	return nil, nil
}

func (c *Voting) PreAuth(ctx context.Context, in *pb.VoterName) (*pb.Challenge, error) {
	name := *in.Name
	for _, vo := range voter_list {
		if name == vo.Name {
			log.Printf("please Give the publickey to Client")
			fmt.Println(vo.Private_key)
			Challenge := make([]byte, 10)
			_, Verr := rand.Read(Challenge)
			if Verr != nil {
				// 处理错误
			}
			vo.Challenge = Challenge
			return &pb.Challenge{Value: vo.Challenge}, nil
		}
	}
	log.Println("Can't Find the Voter~~")
	return nil, nil

}
