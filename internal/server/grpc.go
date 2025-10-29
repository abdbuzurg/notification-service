package server

import (
	pb "asr_leasing_notification/asr_leasing_notification/protos"
	"asr_leasing_notification/internal/subscriber/notifiers"
	"asr_leasing_notification/internal/subscriber/repository"
	"context"
	"database/sql"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedNotificationServiceServer
	dbQuerier *repository.Queries
	notifiers map[repository.NotificationChannel]notifiers.Notifier
}

func NewGRPCServer(
	dbQuerier *repository.Queries,
	notifiers map[repository.NotificationChannel]notifiers.Notifier,
) *Server {
	return &Server{dbQuerier: dbQuerier, notifiers: notifiers}
}

func toProto(n repository.Notification) *pb.Notification {
	proto := &pb.Notification{
		Id:           n.ID,
		UserId:       n.UserID.Int64,
		Channel:      string(n.Channel),
		Recipient:    n.Recipient,
		Subject:      n.Subject.String,
		Body:         n.Body.String,
		Status:       string(n.Status),
		RetryCount:   int32(n.RetryCount.Int32),
		ErrorMessage: n.ErrorMessage.String,
		Source:       n.Source,
		CreatedAt:    timestamppb.New(n.CreatedAt.Time),
	}
	if n.SentAt.Valid {
		proto.SentAt = timestamppb.New(n.SentAt.Time)
	}
	return proto
}

func (s *Server) GetNotificationStatus(ctx context.Context, req *pb.GetStatusReq) (*pb.GetStatusResp, error) {
	notif, err := s.dbQuerier.GetNotificationByID(ctx, req.NotificationId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "notification not found")
	}
	return &pb.GetStatusResp{Notification: toProto(notif)}, nil
}

func (s *Server) ListNotifications(ctx context.Context, req *pb.ListReq) (*pb.ListResp, error) {
	params := sql.NullInt64{Int64: req.UserId, Valid: true}
	notifs, err := s.dbQuerier.ListNotificationsByUserID(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list notifications")
	}
	var pbNotifs []*pb.Notification
	for _, n := range notifs {
		pbNotifs = append(pbNotifs, toProto(n))
	}
	return &pb.ListResp{Notifications: pbNotifs}, nil
}

func (s *Server) ResendNotification(ctx context.Context, req *pb.ResendReq) (*pb.ResendResp, error) {
	dbNotification, err := s.dbQuerier.GetNotificationByID(ctx, req.NotificationId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "notification not found")
	}

	notifier, ok := s.notifiers[dbNotification.Channel]
	if !ok {
		return nil, status.Errorf(codes.Internal, "notifier for channel %s not found", dbNotification.Channel)
	}

	err = notifier.Send(ctx, dbNotification)
	if err != nil {
		s.dbQuerier.UpdateNotificationFailure(ctx, repository.UpdateNotificationFailureParams{
			ID: dbNotification.ID, ErrorMessage: sql.NullString{String: err.Error(), Valid: true},
		})
		return nil, status.Errorf(codes.Internal, "failed to resend notification: %v", err)
	}

	s.dbQuerier.UpdateNotificationSuccess(ctx, dbNotification.ID)
	log.Printf("Successfully RESENT notification via gRPC with ID %d", dbNotification.ID)
	return &pb.ResendResp{Success: true, Message: "notification resent successfully"}, nil
}
