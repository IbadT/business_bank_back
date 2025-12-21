package transportgrpc

import (
	"context"
	"fmt"
	"time"

	balancepb "github.com/IbadT/business_bank_back/pkg/proto"
	baseamountpb "github.com/IbadT/business_bank_back/pkg/proto"
	breakdownpb "github.com/IbadT/business_bank_back/pkg/proto"
	gatewaypb "github.com/IbadT/business_bank_back/pkg/proto"
	generatepb "github.com/IbadT/business_bank_back/pkg/proto"
	holidaypb "github.com/IbadT/business_bank_back/pkg/proto"
	transactionpb "github.com/IbadT/business_bank_back/pkg/proto"
	userpb "github.com/IbadT/business_bank_back/pkg/proto"
	"github.com/IbadT/business_bank_back/services/matematika/internal/domain"
	balanceservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/balance"
	baseamountservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/base"
	breakdownservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/breakdown"
	gatewayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/gateway"
	generatorservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/generator"
	holidayservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/holiday"
	transactionservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/transaction"
	userservice "github.com/IbadT/business_bank_back/services/matematika/internal/service/user"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/helpers"
	"github.com/IbadT/business_bank_back/services/matematika/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	svc                  generatorservice.GeneratorService
	userSvc              userservice.UserService
	transactionSvc       transactionservice.TransactionService
	holidaySvc           holidayservice.HolidayService
	gatewaySvc           gatewayservice.GatewayService
	baseAmountSvc        baseamountservice.BaseAmountService
	breakdownSvc         breakdownservice.BreakdownService
	balanceAdjustmentSvc balanceservice.BalanceAdjustmentService
	generatepb.UnimplementedGenerateServiceServer
	userpb.UnimplementedUserServiceServer
	transactionpb.UnimplementedTransactionServiceServer
	holidaypb.UnimplementedHolidayServiceServer
	gatewaypb.UnimplementedGatewayServiceServer
	baseamountpb.UnimplementedBaseAmountServiceServer
	breakdownpb.UnimplementedBreakdownServiceServer
	balancepb.UnimplementedBalanceServiceServer
}

func NewHandler(
	svc generatorservice.GeneratorService,
	userSvc userservice.UserService,
	transactionSvc transactionservice.TransactionService,
	holidaySvc holidayservice.HolidayService,
	gatewaySvc gatewayservice.GatewayService,
	baseAmountSvc baseamountservice.BaseAmountService,
	breakdownSvc breakdownservice.BreakdownService,
	balanceAdjustmentSvc balanceservice.BalanceAdjustmentService,
) *Handler {
	return &Handler{
		svc:                  svc,
		userSvc:              userSvc,
		transactionSvc:       transactionSvc,
		holidaySvc:           holidaySvc,
		gatewaySvc:           gatewaySvc,
		baseAmountSvc:        baseAmountSvc,
		breakdownSvc:         breakdownSvc,
		balanceAdjustmentSvc: balanceAdjustmentSvc,
	}
}

func (h *Handler) Generate(ctx context.Context, req *generatepb.GenerateRequest) (*generatepb.GenerateResponse, error) {
	op := "grpc.handler.generate"
	log := logger.GetLogger().WithOperation(op).WithFields(logger.Fields{
		"month": req.Month,
		"year":  req.Year,
		"model": req.Model,
	})
	
	log.Info("Generating transactions")
	
	dtoReq := &dto.GenerateRequest{
		Month:                int(req.Month),
		Year:                 int(req.Year),
		Turnover:             req.Turnover,
		DesiredProfitPercent: req.DesiredProfitPercent,
		Model:                req.Model,
		InitialBalance:       req.InitialBalance,
		ScaleFactor:          int(req.ScaleFactor),
	}

	if req.CustomData != nil {
		dtoReq.CustomData = &dto.CustomData{
			CompanyInfo: dto.CompanyInfo{
				OwnerName:   req.CustomData.CompanyInfo.OwnerName,
				CompanyName: req.CustomData.CompanyInfo.CompanyName,
			},
			CustomCustomers: req.CustomData.CustomCustomers,
		}

		for _, mt := range req.CustomData.ManualTransactions {
			date, err := time.Parse("2006-01-02", mt.TransactionDate)
			if err != nil {
				log.Error(err, "Invalid transaction date format: %s", mt.TransactionDate)
				return nil, status.Errorf(codes.InvalidArgument, "invalid transaction date: %v", err)
			}
			dtoReq.CustomData.ManualTransactions = append(dtoReq.CustomData.ManualTransactions, dto.ManualTransaction{
				TransactionDate: date,
				Type:            mt.Type,
				Category:        mt.Category,
				Method:          mt.Method,
				Amount:          mt.Amount,
			})
		}

		for _, cc := range req.CustomData.CustomContractors {
			dtoReq.CustomData.CustomContractors = append(dtoReq.CustomData.CustomContractors, dto.CustomContractor{
				TransactionType: cc.TransactionType,
				Name:            cc.Name,
			})
		}
	}

	// Получаем userID из metadata
	userIDStr, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	userID := userIDStr
	log = log.WithFields(logger.Fields{"user_id": userID})

	dtoResp, err := h.svc.GenerateTransactions(dtoReq, &userID)
	if err != nil {
		log.Error(err, "Failed to generate transactions")
		return nil, status.Errorf(codes.Internal, "failed to generate transactions: %v", err)
	}
	
	log.WithFields(logger.Fields{
		"request_id": dtoResp.RequestID,
		"total_transactions": dtoResp.TransactionCounts.Total,
	}).Success("Transactions generated successfully")

	protoResp := &generatepb.GenerateResponse{
		RequestId: dtoResp.RequestID,
		FinancialSummary: &generatepb.FinancialSummary{
			InitialBalance: dtoResp.FinancialSummary.InitialBalance,
			FinalBalance:   dtoResp.FinancialSummary.FinalBalance,
			TotalRevenue:   dtoResp.FinancialSummary.TotalRevenue,
			TotalExpenses:  dtoResp.FinancialSummary.TotalExpenses,
			NetProfit:      dtoResp.FinancialSummary.NetProfit,
		},
		ForwardingInfo: &generatepb.ForwardingInfo{
			AssociatedCard:  dtoResp.ForwardingInfo.AssociatedCard,
			OwnerName:       dtoResp.ForwardingInfo.OwnerName,
			CompanyName:     dtoResp.ForwardingInfo.CompanyName,
			CustomCustomers: dtoResp.ForwardingInfo.CustomCustomers,
		},
		RevenueBreakdown: &generatepb.RevenueBreakdown{
			TotalAch:     dtoResp.RevenueBreakdown.TotalAch,
			TotalWire:    dtoResp.RevenueBreakdown.TotalWire,
			TotalZelle:   dtoResp.RevenueBreakdown.TotalZelle,
			TotalGateway: dtoResp.RevenueBreakdown.TotalGateway,
			TotalOther:   dtoResp.RevenueBreakdown.TotalOther,
		},
		ExpensesBreakdown: &generatepb.ExpensesBreakdown{
			ByCard:    dtoResp.ExpensesBreakdown.ByCard,
			ByAccount: dtoResp.ExpensesBreakdown.ByAccount,
		},
		TransactionCounts: &generatepb.TransactionCounts{
			TotalTransactions: int32(dtoResp.TransactionCounts.Total),
			TotalIncome:       0, // TODO: добавить в TransactionCounts
			TotalExpenses:     0, // TODO: добавить в TransactionCounts
		},
	}

	for _, tx := range dtoResp.Transactions {
		protoTx := &generatepb.GeneratedTransaction{
			Id:              tx.TransactionID,
			RequestId:       dtoResp.RequestID,
			TransactionId:   tx.TransactionID,
			TransactionDate: tx.TransactionDate.Format("2006-01-02"),
			PostingDate:     tx.PostingDate,
			Type:            tx.Type,
			Category:        tx.Category,
			Method:          tx.Method,
			Amount:          tx.Amount,
			BalanceAfter:    tx.BalanceAfter,
			IsManual:        tx.IsManual,
		}
		if tx.CalculationDetails != nil {
			protoTx.CalculationDetails = make(map[string]string)
			for k, v := range tx.CalculationDetails {
				protoTx.CalculationDetails[k] = fmt.Sprintf("%v", v)
			}
		}
		protoResp.Transactions = append(protoResp.Transactions, protoTx)
	}

	for _, db := range dtoResp.DailyClosingBalances {
		protoResp.DailyBalances = append(protoResp.DailyBalances, &generatepb.DailyBalance{
			Date:    db.Date,
			Balance: db.Balance,
		})
	}

	for _, cc := range dtoResp.ForwardingInfo.CustomContractors {
		protoResp.ForwardingInfo.CustomContractors = append(protoResp.ForwardingInfo.CustomContractors, &generatepb.CustomContractor{
			TransactionType: cc.TransactionType,
			Name:            cc.Name,
		})
	}

	return protoResp, nil
}

func (h *Handler) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	op := "grpc.handler.login"
	log := logger.GetLogger().WithOperation(op)
	
	log.Info("Login attempt for email: %s", req.Email)
	
	token, err := h.userSvc.Login(req.Email, req.Password)
	if err != nil {
		log.Error(err, "Login failed for email: %s", req.Email)
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}
	
	log.WithFields(logger.Fields{"email": req.Email}).Success("User logged in successfully")

	return &userpb.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (h *Handler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	op := "grpc.handler.register"
	log := logger.GetLogger().WithOperation(op)
	
	log.Info("Registration attempt for email: %s", req.Email)
	
	token, err := h.userSvc.Register(req.Email, req.Password)
	if err != nil {
		log.Error(err, "Registration failed for email: %s", req.Email)
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}
	
	log.WithFields(logger.Fields{"email": req.Email}).Success("User registered successfully")

	return &userpb.RegisterResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (h *Handler) AssociatedCard(ctx context.Context, req *userpb.AssociatedCardRequest) (*userpb.AssociatedCardResponse, error) {
	op := "grpc.handler.associatedCard"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{
		"user_id": userID,
		"card":    req.AssociatedCard,
	})
	log.Info("Saving associated card")
	
	err = h.userSvc.SaveAssociatedCard(userID, req.AssociatedCard)
	if err != nil {
		log.Error(err, "Failed to save associated card")
		return nil, status.Errorf(codes.Internal, "failed to save associated card: %v", err)
	}
	
	log.Success("Associated card saved successfully")

	return &userpb.AssociatedCardResponse{
		Message: "Associated card saved successfully",
		Code:    200,
	}, nil
}

func (h *Handler) CreateTransaction(ctx context.Context, req *transactionpb.CreateTransactionRequest) (*transactionpb.CreateTransactionResponse, error) {
	op := "grpc.handler.createTransaction"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{
		"request_id": req.RequestId,
		"type":       req.Type,
		"method":     req.Method,
		"amount":     req.Amount,
	})
	log.Info("Creating transaction")
	
	dtoReq := &dto.CreateTransactionRequest{
		RequestID:       req.RequestId,
		TransactionDate: req.TransactionDate,
		PostingDate:     req.PostingDate,
		Type:            req.Type,
		Category:        req.Category,
		Method:          req.Method,
		Amount:          req.Amount,
	}

	err := h.transactionSvc.CreateTransaction(dtoReq)
	if err != nil {
		log.Error(err, "Failed to create transaction")
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}
	
	log.Success("Transaction created successfully")

	return &transactionpb.CreateTransactionResponse{
		Message: "Transaction created successfully",
		Code:    200,
	}, nil
}

func (h *Handler) CreateBatchTransactions(ctx context.Context, req *transactionpb.CreateBatchTransactionsRequest) (*transactionpb.CreateBatchTransactionsResponse, error) {
	op := "grpc.handler.createBatchTransactions"
	log := logger.GetLogger().WithOperation(op)
	
	log.Warn("Batch transactions not implemented")
	return nil, status.Errorf(codes.Unimplemented, "batch transactions not implemented in proto")
}

func (h *Handler) GetTransactionsCount(ctx context.Context, req *transactionpb.GetTransactionsCountRequest) (*transactionpb.GetTransactionsCountResponse, error) {
	op := "grpc.handler.getTransactionsCount"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"request_id": req.RequestId})
	log.Info("Getting transactions count")
	
	count, err := h.transactionSvc.GetCountByRequestID(req.RequestId)
	if err != nil {
		log.Error(err, "Failed to get transactions count")
		return nil, status.Errorf(codes.Internal, "failed to get transactions count: %v", err)
	}
	
	log.WithFields(logger.Fields{"count": count}).Success("Transactions count retrieved")

	return &transactionpb.GetTransactionsCountResponse{
		Count: count,
		Code:  200,
	}, nil
}

func (h *Handler) GetTransactionsByTypeAndRequestID(ctx context.Context, req *transactionpb.GetTransactionsByTypeAndRequestIDRequest) (*transactionpb.GetTransactionsByTypeAndRequestIDResponse, error) {
	op := "grpc.handler.getTransactionsByTypeAndRequestID"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{
		"request_id": req.RequestId,
		"type":       req.Type,
	})
	log.Info("Getting transactions by type")
	
	transactions, err := h.transactionSvc.GetByTypeAndRequestID(req.Type, req.RequestId)
	if err != nil {
		log.Error(err, "Failed to get transactions by type")
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}
	
	log.WithFields(logger.Fields{"count": len(protoTxs)}).Success("Transactions retrieved by type")

	return &transactionpb.GetTransactionsByTypeAndRequestIDResponse{
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

func (h *Handler) GetTransactionsByMethodAndRequestID(ctx context.Context, req *transactionpb.GetTransactionsByMethodAndRequestIDRequest) (*transactionpb.GetTransactionsByMethodAndRequestIDResponse, error) {
	op := "grpc.handler.getTransactionsByMethodAndRequestID"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{
		"request_id": req.RequestId,
		"method":    req.Method,
	})
	log.Info("Getting transactions by method")
	
	transactions, err := h.transactionSvc.GetByMethodAndRequestID(req.Method, req.RequestId)
	if err != nil {
		log.Error(err, "Failed to get transactions by method")
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}
	
	log.WithFields(logger.Fields{"count": len(protoTxs)}).Success("Transactions retrieved by method")

	return &transactionpb.GetTransactionsByMethodAndRequestIDResponse{
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

func (h *Handler) GetTransactionsByRequestID(ctx context.Context, req *transactionpb.GetTransactionsByRequestIDRequest) (*transactionpb.GetTransactionsByRequestIDResponse, error) {
	op := "grpc.handler.getTransactionsByRequestID"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"request_id": req.RequestId})
	log.Info("Getting transactions by request ID")
	
	transactions, err := h.transactionSvc.GetByRequestID(req.RequestId)
	if err != nil {
		log.Error(err, "Failed to get transactions by request ID")
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}
	
	log.WithFields(logger.Fields{"count": len(protoTxs)}).Success("Transactions retrieved by request ID")

	return &transactionpb.GetTransactionsByRequestIDResponse{
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

func (h *Handler) AddHoliday(ctx context.Context, req *holidaypb.AddHolidayRequest) (*holidaypb.AddHolidayResponse, error) {
	op := "grpc.handler.addHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{
		"name":    req.Holiday.Name,
		"date":    req.Holiday.HolidayDate,
		"country": req.Holiday.Country,
	})
	log.Info("Adding holiday")
	
	date, err := time.Parse("2006-01-02", req.Holiday.HolidayDate)
	if err != nil {
		log.Error(err, "Invalid date format: %s", req.Holiday.HolidayDate)
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	err = h.holidaySvc.AddHoliday(date, req.Holiday.Name, req.Holiday.Country)
	if err != nil {
		log.Error(err, "Failed to add holiday")
		return nil, status.Errorf(codes.Internal, "failed to add holiday: %v", err)
	}
	
	log.Success("Holiday added successfully")

	return &holidaypb.AddHolidayResponse{
		Message: "Holiday added successfully",
		Code:    200,
	}, nil
}

func (h *Handler) GetHolidays(ctx context.Context, req *holidaypb.GetHolidaysRequest) (*holidaypb.GetHolidaysResponse, error) {
	op := "grpc.handler.getHolidays"
	log := logger.GetLogger().WithOperation(op)
	
	log.Info("Getting holidays for year: %s", req.Year)
	
	year, err := time.Parse("2006", req.Year)
	if err != nil {
		log.Error(err, "Invalid year format: %s", req.Year)
		return nil, status.Errorf(codes.InvalidArgument, "invalid year format: %v", err)
	}

	holidays, err := h.holidaySvc.GetHolidays(year)
	if err != nil {
		log.Error(err, "Failed to get holidays")
		return nil, status.Errorf(codes.Internal, "failed to get holidays: %v", err)
	}

	protoHolidays := make([]*holidaypb.Holiday, 0, len(holidays))
	for _, h := range holidays {
		protoHolidays = append(protoHolidays, &holidaypb.Holiday{
			Id:          "", // TODO: добавить ID в domain.Holiday
			HolidayDate: h.HolidayDate,
			Name:        h.Name,
			Country:     h.Country,
		})
	}
	
	log.WithFields(logger.Fields{
		"year":  req.Year,
		"count": len(protoHolidays),
	}).Success("Holidays retrieved")

	return &holidaypb.GetHolidaysResponse{
		Holidays: protoHolidays,
		Year:     req.Year,
	}, nil
}

func (h *Handler) CheckIsHoliday(ctx context.Context, req *holidaypb.CheckIsHolidayRequest) (*holidaypb.CheckIsHolidayResponse, error) {
	op := "grpc.handler.checkIsHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	log.Info("Checking if date is holiday: %s", req.Date)
	
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		log.Error(err, "Invalid date format: %s", req.Date)
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	isHoliday := h.holidaySvc.IsHoliday(date)
	
	log.WithFields(logger.Fields{
		"date":      req.Date,
		"is_holiday": isHoliday,
	}).Info("Holiday check completed")

	return &holidaypb.CheckIsHolidayResponse{
		IsHoliday: isHoliday,
		Date:      req.Date,
	}, nil
}

func (h *Handler) UpdateHoliday(ctx context.Context, req *holidaypb.UpdateHolidayRequest) (*holidaypb.UpdateHolidayResponse, error) {
	op := "grpc.handler.updateHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"holiday_id": req.Holiday.Id})
	log.Info("Updating holiday")
	
	id, err := helpers.ParseHolidayID(req.Holiday.Id)
	if err != nil {
		log.Error(err, "Invalid holiday id: %s", req.Holiday.Id)
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	date, err := time.Parse("2006-01-02", req.Holiday.HolidayDate)
	if err != nil {
		log.Error(err, "Invalid date format: %s", req.Holiday.HolidayDate)
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	err = h.holidaySvc.UpdateHoliday(id, date, req.Holiday.Name, req.Holiday.Country)
	if err != nil {
		log.Error(err, "Failed to update holiday")
		return nil, status.Errorf(codes.Internal, "failed to update holiday: %v", err)
	}
	
	log.Success("Holiday updated successfully")

	return &holidaypb.UpdateHolidayResponse{
		Message: "Holiday updated successfully",
		Code:    200,
	}, nil
}

func (h *Handler) DeleteHoliday(ctx context.Context, req *holidaypb.DeleteHolidayRequest) (*holidaypb.DeleteHolidayResponse, error) {
	op := "grpc.handler.deleteHoliday"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"holiday_id": req.Id})
	log.Info("Deleting holiday")
	
	id, err := helpers.ParseHolidayID(req.Id)
	if err != nil {
		log.Error(err, "Invalid holiday id: %s", req.Id)
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	err = h.holidaySvc.DeleteHoliday(id)
	if err != nil {
		log.Error(err, "Failed to delete holiday")
		return nil, status.Errorf(codes.Internal, "failed to delete holiday: %v", err)
	}
	
	log.Success("Holiday deleted successfully")

	return &holidaypb.DeleteHolidayResponse{
		Message: "Holiday deleted successfully",
		Code:    200,
	}, nil
}

func (h *Handler) GetB2CGateways(ctx context.Context, req *gatewaypb.GetB2CGatewaysRequest) (*gatewaypb.GetB2CGatewaysResponse, error) {
	op := "grpc.handler.getB2CGateways"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting B2C gateways")
	
	userIDUUID, err := helpers.ParseUserID(userID)
	if err != nil {
		log.Error(err, "Invalid user id: %s", userID)
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	gateway, err := h.gatewaySvc.GetB2CGateways(userIDUUID)
	if err != nil {
		log.Error(err, "Failed to get gateway")
		return nil, status.Errorf(codes.Internal, "failed to get gateway: %v", err)
	}
	
	log.WithFields(logger.Fields{
		"gateway_id": gateway.ID,
		"gateway_name": gateway.Name,
	}).Success("B2C gateway retrieved")

	return &gatewaypb.GetB2CGatewaysResponse{
		Gateway: &gatewaypb.Gateway{
			Id:   gateway.ID,
			Name: gateway.Name,
		},
		Code: 200,
	}, nil
}

func (h *Handler) UpdateB2CGateways(ctx context.Context, req *gatewaypb.UpdateB2CGatewaysRequest) (*gatewaypb.UpdateB2CGatewaysResponse, error) {
	op := "grpc.handler.updateB2CGateways"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{
		"user_id":   userID,
		"gateway_id": req.GatewayId,
	})
	log.Info("Updating B2C gateway")
	
	userIDUUID, err := helpers.ParseUserID(userID)
	if err != nil {
		log.Error(err, "Invalid user id: %s", userID)
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	err = h.gatewaySvc.SaveB2CGateways(userIDUUID, req.GatewayId)
	if err != nil {
		log.Error(err, "Failed to update gateway")
		return nil, status.Errorf(codes.Internal, "failed to update gateway: %v", err)
	}
	
	log.Success("B2C gateway updated successfully")

	return &gatewaypb.UpdateB2CGatewaysResponse{
		Message: "Gateway updated successfully",
		Code:    200,
	}, nil
}

func (h *Handler) DeleteB2CGateways(ctx context.Context, req *gatewaypb.DeleteB2CGatewaysRequest) (*gatewaypb.DeleteB2CGatewaysResponse, error) {
	op := "grpc.handler.deleteB2CGateways"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{"user_id": userID})
	log.Info("Deleting B2C gateway")
	
	userIDUUID, err := helpers.ParseUserID(userID)
	if err != nil {
		log.Error(err, "Invalid user id: %s", userID)
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	err = h.gatewaySvc.DeleteB2CGateways(userIDUUID)
	if err != nil {
		log.Error(err, "Failed to delete gateway")
		return nil, status.Errorf(codes.Internal, "failed to delete gateway: %v", err)
	}
	
	log.Success("B2C gateway deleted successfully")

	return &gatewaypb.DeleteB2CGatewaysResponse{
		Message: "Gateway deleted successfully",
		Code:    200,
	}, nil
}

func (h *Handler) GetBaseAmount(ctx context.Context, req *baseamountpb.GetBaseAmountRequest) (*baseamountpb.GetBaseAmountResponse, error) {
	op := "grpc.handler.getBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{"user_id": userID})
	log.Info("Getting base amounts")

	baseAmounts, err := h.baseAmountSvc.GetBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to get base amount")
		return nil, status.Errorf(codes.Internal, "failed to get base amount: %v", err)
	}
	
	log.WithFields(logger.Fields{
		"mobile":    baseAmounts.MobileBaseAmount.Amount,
		"utilities": baseAmounts.UtilitiesBaseAmount.Amount,
		"leasing":   baseAmounts.LeasingBaseAmount.Amount,
	}).Success("Base amounts retrieved")

	return &baseamountpb.GetBaseAmountResponse{
		UserId:              userID,
		MobileBaseAmount:    baseAmounts.MobileBaseAmount.Amount,
		UtilitiesBaseAmount: baseAmounts.UtilitiesBaseAmount.Amount,
		LeasingBaseAmount:   baseAmounts.LeasingBaseAmount.Amount,
		Code:                200,
	}, nil
}

func (h *Handler) CalculateMobileAmount(ctx context.Context, req *baseamountpb.CalculateMobileAmountRequest) (*baseamountpb.CalculateMobileAmountResponse, error) {
	op := "grpc.handler.calculateMobileAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{
		"user_id":      userID,
		"is_first_month": req.IsFirstMonth,
		"month":        req.Month,
	})
	log.Info("Calculating mobile amount")

	amount, err := h.baseAmountSvc.CalculateMobileAmount(userID, req.IsFirstMonth, req.Month)
	if err != nil {
		log.Error(err, "Failed to calculate mobile amount")
		return nil, status.Errorf(codes.Internal, "failed to calculate mobile amount: %v", err)
	}
	
	log.WithFields(logger.Fields{"amount": amount}).Success("Mobile amount calculated")

	return &baseamountpb.CalculateMobileAmountResponse{
		UserId:       userID,
		Amount:       amount,
		IsFirstMonth: req.IsFirstMonth,
		Code:         200,
	}, nil
}

func (h *Handler) CalculateUtilitiesAmount(ctx context.Context, req *baseamountpb.CalculateUtilitiesAmountRequest) (*baseamountpb.CalculateUtilitiesAmountResponse, error) {
	op := "grpc.handler.calculateUtilitiesAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{
		"user_id":      userID,
		"is_first_month": req.IsFirstMonth,
		"month":        req.Month,
	})
	log.Info("Calculating utilities amount")

	amount, err := h.baseAmountSvc.CalculateUtilitiesAmount(userID, req.IsFirstMonth, req.Month)
	if err != nil {
		log.Error(err, "Failed to calculate utilities amount")
		return nil, status.Errorf(codes.Internal, "failed to calculate utilities amount: %v", err)
	}
	
	log.WithFields(logger.Fields{"amount": amount}).Success("Utilities amount calculated")

	return &baseamountpb.CalculateUtilitiesAmountResponse{
		UserId:       userID,
		Amount:       amount,
		IsFirstMonth: req.IsFirstMonth,
		Code:         200,
	}, nil
}

func (h *Handler) CalculateLeasingAmount(ctx context.Context, req *baseamountpb.CalculateLeasingAmountRequest) (*baseamountpb.CalculateLeasingAmountResponse, error) {
	op := "grpc.handler.calculateLeasingAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{
		"user_id":      userID,
		"is_first_month": req.IsFirstMonth,
		"month":        req.Month,
		"turnover":     req.Turnover,
	})
	log.Info("Calculating leasing amount")
	
	amount, err := h.baseAmountSvc.CalculateLeasingAmount(userID, req.Turnover, req.IsFirstMonth, req.Month)
	if err != nil {
		log.Error(err, "Failed to calculate leasing amount")
		return nil, status.Errorf(codes.Internal, "failed to calculate leasing amount: %v", err)
	}
	
	log.WithFields(logger.Fields{"amount": amount}).Success("Leasing amount calculated")

	return &baseamountpb.CalculateLeasingAmountResponse{
		UserId:       userID,
		Amount:       amount,
		IsFirstMonth: req.IsFirstMonth,
		Turnover:     req.Turnover,
		Code:         200,
	}, nil
}

func (h *Handler) ResetMobileBaseAmount(ctx context.Context, req *baseamountpb.ResetMobileBaseAmountRequest) (*baseamountpb.ResetMobileBaseAmountResponse, error) {
	op := "grpc.handler.resetMobileBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{"user_id": userID})
	log.Info("Resetting mobile base amount")
	
	err = h.baseAmountSvc.DeleteMobileBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to reset mobile base amount")
		return nil, status.Errorf(codes.Internal, "failed to reset mobile base amount: %v", err)
	}
	
	log.Success("Mobile base amount reset successfully")

	return &baseamountpb.ResetMobileBaseAmountResponse{
		Message: "Mobile base amount reset successfully",
		Code:    200,
	}, nil
}

func (h *Handler) ResetUtilitiesBaseAmount(ctx context.Context, req *baseamountpb.ResetUtilitiesBaseAmountRequest) (*baseamountpb.ResetUtilitiesBaseAmountResponse, error) {
	op := "grpc.handler.resetUtilitiesBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{"user_id": userID})
	log.Info("Resetting utilities base amount")
	
	err = h.baseAmountSvc.DeleteUtilitiesBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to reset utilities base amount")
		return nil, status.Errorf(codes.Internal, "failed to reset utilities base amount: %v", err)
	}
	
	log.Success("Utilities base amount reset successfully", logger.Fields{"user_id": userID})

	return &baseamountpb.ResetUtilitiesBaseAmountResponse{
		Message: "Utilities base amount reset successfully",
		Code:    200,
	}, nil
}

func (h *Handler) ResetLeasingBaseAmount(ctx context.Context, req *baseamountpb.ResetLeasingBaseAmountRequest) (*baseamountpb.ResetLeasingBaseAmountResponse, error) {
	op := "grpc.handler.resetLeasingBaseAmount"
	log := logger.GetLogger().WithOperation(op)
	
	userID, err := getUserIDFromMetadata(ctx)
	if err != nil {
		log.Error(err, "Failed to get user_id from metadata")
		return nil, err
	}
	log = log.WithFields(logger.Fields{"user_id": userID})
	log.Info("Resetting leasing base amount")
	
	err = h.baseAmountSvc.DeleteLeasingBaseAmount(userID)
	if err != nil {
		log.Error(err, "Failed to reset leasing base amount")
		return nil, status.Errorf(codes.Internal, "failed to reset leasing base amount: %v", err)
	}
	
	log.Success("Leasing base amount reset successfully")

	return &baseamountpb.ResetLeasingBaseAmountResponse{
		Message: "Leasing base amount reset successfully",
		Code:    200,
	}, nil
}

func (h *Handler) CalculateRevenueBreakdown(ctx context.Context, req *breakdownpb.CalculateRevenueBreakdownRequest) (*breakdownpb.CalculateRevenueBreakdownResponse, error) {
	op := "grpc.handler.calculateRevenueBreakdown"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"request_id": req.RequestId})
	log.Info("Calculating revenue breakdown")
	
	breakdown, err := h.breakdownSvc.GetRevenueBreakdown(req.RequestId)
	if err != nil {
		log.Error(err, "Failed to calculate revenue breakdown")
		return nil, status.Errorf(codes.Internal, "failed to calculate revenue breakdown: %v", err)
	}
	
	log.WithFields(logger.Fields{
		"total_ach":     breakdown.TotalAch,
		"total_wire":    breakdown.TotalWire,
		"total_zelle":   breakdown.TotalZelle,
		"total_gateway": breakdown.TotalGateway,
	}).Success("Revenue breakdown calculated")

	return &breakdownpb.CalculateRevenueBreakdownResponse{
		RequestId: req.RequestId,
		RevenueBreakdown: &breakdownpb.RevenueBreakdown{
			TotalAch:     breakdown.TotalAch,
			TotalWire:    breakdown.TotalWire,
			TotalZelle:   breakdown.TotalZelle,
			TotalGateway: breakdown.TotalGateway,
			TotalOther:   breakdown.TotalOther,
		},
		Code: 200,
	}, nil
}

func (h *Handler) CalculateExpensesBreakdown(ctx context.Context, req *breakdownpb.CalculateExpensesBreakdownRequest) (*breakdownpb.CalculateExpensesBreakdownResponse, error) {
	op := "grpc.handler.calculateExpensesBreakdown"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"request_id": req.RequestId})
	log.Info("Calculating expenses breakdown")
	
	breakdown, err := h.breakdownSvc.GetExpensesBreakdown(req.RequestId)
	if err != nil {
		log.Error(err, "Failed to calculate expenses breakdown")
		return nil, status.Errorf(codes.Internal, "failed to calculate expenses breakdown: %v", err)
	}
	
	log.WithFields(logger.Fields{
		"by_card":    breakdown.ByCard,
		"by_account": breakdown.ByAccount,
	}).Success("Expenses breakdown calculated")

	return &breakdownpb.CalculateExpensesBreakdownResponse{
		RequestId: req.RequestId,
		ExpensesBreakdown: &breakdownpb.ExpensesBreakdown{
			ByCard:    breakdown.ByCard,
			ByAccount: breakdown.ByAccount,
		},
		Code: 200,
	}, nil
}

func (h *Handler) ValidateBalance(ctx context.Context, req *balancepb.ValidateBalanceRequest) (*balancepb.ValidateBalanceResponse, error) {
	op := "grpc.handler.validateBalance"
	log := logger.GetLogger().WithOperation(op)
	
	log = log.WithFields(logger.Fields{"request_id": req.RequestId})
	log.Info("Validating balance")
	
	result, err := h.balanceAdjustmentSvc.ValidateBalance(req.RequestId)
	if err != nil {
		log.Error(err, "Failed to validate balance", logger.Fields{"request_id": req.RequestId})
		return nil, status.Errorf(codes.Internal, "failed to validate balance: %v", err)
	}

	issues := make([]*balancepb.BalanceIssue, 0, len(result.Issues))
	for _, issue := range result.Issues {
		issues = append(issues, &balancepb.BalanceIssue{
			TransactionId:    issue.TransactionID,
			Date:             issue.Date,
			RequiredBalance:  issue.RequiredBalance,
			AvailableBalance: issue.AvailableBalance,
			Shortage:         issue.Shortage,
			ActionTaken:      issue.ActionTaken,
			NewDate:          issue.NewDate,
		})
	}
	
	log.WithFields(logger.Fields{
		"is_valid":   result.IsValid,
		"issues_count": len(issues),
	}).Success("Balance validation completed")

	return &balancepb.ValidateBalanceResponse{
		RequestId: req.RequestId,
		IsValid:   result.IsValid,
		Issues:    issues,
		Code:      200,
	}, nil
}

func (h *Handler) GetBalanceAdjustment(ctx context.Context, req *balancepb.GetBalanceAdjustmentRequest) (*balancepb.GetBalanceAdjustmentResponse, error) {
	op := "grpc.handler.getBalanceAdjustment"
	log := logger.GetLogger().WithOperation(op)
	
	log.Info("Getting balance adjustment", logger.Fields{"request_id": req.RequestId})
	
	transactions, err := h.balanceAdjustmentSvc.GetAdjustedTransactions(req.RequestId)
	if err != nil {
		log.Error(err, "Failed to get balance adjustment")
		return nil, status.Errorf(codes.Internal, "failed to get balance adjustment: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}
	
	log.WithFields(logger.Fields{"transactions_count": len(protoTxs)}).Success("Balance adjustment retrieved")

	return &balancepb.GetBalanceAdjustmentResponse{
		RequestId:    req.RequestId,
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

// TODO: вынести в отдельный файл
func domainToProtoTransaction(tx *domain.GeneratedTransaction, requestID string) *generatepb.GeneratedTransaction {
	protoTx := &generatepb.GeneratedTransaction{
		Id:              tx.ID.String(),
		RequestId:       requestID,
		TransactionId:   tx.TransactionID,
		TransactionDate: tx.TransactionDate.Format("2006-01-02"),
		PostingDate:     tx.PostingDate.Format("2006-01-02"),
		Type:            tx.Type,
		Category:        tx.Category,
		Method:          tx.Method,
		Amount:          tx.Amount,
		IsManual:        tx.IsManual,
	}

	if tx.BalanceAfter != nil {
		protoTx.BalanceAfter = *tx.BalanceAfter
	}
	if tx.SortOrder != nil {
		protoTx.SortOrder = int32(*tx.SortOrder)
	}
	if tx.CalculationDetails != nil {
		protoTx.CalculationDetails = make(map[string]string)
		for k, v := range tx.CalculationDetails {
			protoTx.CalculationDetails[k] = fmt.Sprintf("%v", v)
		}
	}

	return protoTx
}
