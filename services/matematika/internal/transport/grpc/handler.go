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
	"github.com/IbadT/business_bank_back/services/matematika/internal/service"
	"github.com/IbadT/business_bank_back/services/matematika/internal/transport/http/dto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	svc                  service.GeneratorService
	userSvc              service.UserService
	transactionSvc       service.TransactionService
	holidaySvc           service.HolidayService
	gatewaySvc           service.GatewayService
	baseAmountSvc        service.BaseAmountService
	breakdownSvc         service.BreakdownService
	balanceAdjustmentSvc service.BalanceAdjustmentService
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
	svc service.GeneratorService,
	userSvc service.UserService,
	transactionSvc service.TransactionService,
	holidaySvc service.HolidayService,
	gatewaySvc service.GatewayService,
	baseAmountSvc service.BaseAmountService,
	breakdownSvc service.BreakdownService,
	balanceAdjustmentSvc service.BalanceAdjustmentService,
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

	// TODO: получить userID из контекста (из метаданных gRPC)
	userID := ""

	dtoResp, err := h.svc.GenerateTransactions(dtoReq, &userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate transactions: %v", err)
	}

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
	token, err := h.userSvc.Login(req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	return &userpb.LoginResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (h *Handler) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.RegisterResponse, error) {
	token, err := h.userSvc.Register(req.Email, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &userpb.RegisterResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}, nil
}

func (h *Handler) AssociatedCard(ctx context.Context, req *userpb.AssociatedCardRequest) (*userpb.AssociatedCardResponse, error) {
	userID := ""
	err := h.userSvc.SaveAssociatedCard(userID, req.AssociatedCard)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save associated card: %v", err)
	}

	return &userpb.AssociatedCardResponse{
		Message: "Associated card saved successfully",
		Code:    200,
	}, nil
}

func (h *Handler) CreateTransaction(ctx context.Context, req *transactionpb.CreateTransactionRequest) (*transactionpb.CreateTransactionResponse, error) {
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
		return nil, status.Errorf(codes.Internal, "failed to create transaction: %v", err)
	}

	return &transactionpb.CreateTransactionResponse{
		Message: "Transaction created successfully",
		Code:    200,
	}, nil
}

func (h *Handler) CreateBatchTransactions(ctx context.Context, req *transactionpb.CreateBatchTransactionsRequest) (*transactionpb.CreateBatchTransactionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "batch transactions not implemented in proto")
}

func (h *Handler) GetTransactionsCount(ctx context.Context, req *transactionpb.GetTransactionsCountRequest) (*transactionpb.GetTransactionsCountResponse, error) {
	count, err := h.transactionSvc.GetCountByRequestID(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions count: %v", err)
	}

	return &transactionpb.GetTransactionsCountResponse{
		Count: count,
		Code:  200,
	}, nil
}

func (h *Handler) GetTransactionsByTypeAndRequestID(ctx context.Context, req *transactionpb.GetTransactionsByTypeAndRequestIDRequest) (*transactionpb.GetTransactionsByTypeAndRequestIDResponse, error) {
	transactions, err := h.transactionSvc.GetByTypeAndRequestID(req.Type, req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}

	return &transactionpb.GetTransactionsByTypeAndRequestIDResponse{
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

func (h *Handler) GetTransactionsByMethodAndRequestID(ctx context.Context, req *transactionpb.GetTransactionsByMethodAndRequestIDRequest) (*transactionpb.GetTransactionsByMethodAndRequestIDResponse, error) {
	transactions, err := h.transactionSvc.GetByMethodAndRequestID(req.Method, req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}

	return &transactionpb.GetTransactionsByMethodAndRequestIDResponse{
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

func (h *Handler) GetTransactionsByRequestID(ctx context.Context, req *transactionpb.GetTransactionsByRequestIDRequest) (*transactionpb.GetTransactionsByRequestIDResponse, error) {
	transactions, err := h.transactionSvc.GetByRequestID(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get transactions: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}

	return &transactionpb.GetTransactionsByRequestIDResponse{
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

func (h *Handler) AddHoliday(ctx context.Context, req *holidaypb.AddHolidayRequest) (*holidaypb.AddHolidayResponse, error) {
	date, err := time.Parse("2006-01-02", req.Holiday.HolidayDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	err = h.holidaySvc.AddHoliday(date, req.Holiday.Name, req.Holiday.Country)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add holiday: %v", err)
	}

	return &holidaypb.AddHolidayResponse{
		Message: "Holiday added successfully",
		Code:    200,
	}, nil
}

func (h *Handler) GetHolidays(ctx context.Context, req *holidaypb.GetHolidaysRequest) (*holidaypb.GetHolidaysResponse, error) {
	year, err := time.Parse("2006", req.Year)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid year format: %v", err)
	}

	holidays, err := h.holidaySvc.GetHolidays(year)
	if err != nil {
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

	return &holidaypb.GetHolidaysResponse{
		Holidays: protoHolidays,
		Year:     req.Year,
	}, nil
}

func (h *Handler) CheckIsHoliday(ctx context.Context, req *holidaypb.CheckIsHolidayRequest) (*holidaypb.CheckIsHolidayResponse, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	isHoliday := h.holidaySvc.IsHoliday(date)

	return &holidaypb.CheckIsHolidayResponse{
		IsHoliday: isHoliday,
		Date:      req.Date,
	}, nil
}

func (h *Handler) UpdateHoliday(ctx context.Context, req *holidaypb.UpdateHolidayRequest) (*holidaypb.UpdateHolidayResponse, error) {
	id, err := uuid.Parse(req.Holiday.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid holiday id: %v", err)
	}

	date, err := time.Parse("2006-01-02", req.Holiday.HolidayDate)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date format: %v", err)
	}

	err = h.holidaySvc.UpdateHoliday(id, date, req.Holiday.Name, req.Holiday.Country)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update holiday: %v", err)
	}

	return &holidaypb.UpdateHolidayResponse{
		Message: "Holiday updated successfully",
		Code:    200,
	}, nil
}

func (h *Handler) DeleteHoliday(ctx context.Context, req *holidaypb.DeleteHolidayRequest) (*holidaypb.DeleteHolidayResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid holiday id: %v", err)
	}

	err = h.holidaySvc.DeleteHoliday(id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete holiday: %v", err)
	}

	return &holidaypb.DeleteHolidayResponse{
		Message: "Holiday deleted successfully",
		Code:    200,
	}, nil
}

func (h *Handler) GetB2CGateways(ctx context.Context, req *gatewaypb.GetB2CGatewaysRequest) (*gatewaypb.GetB2CGatewaysResponse, error) {
	userID := uuid.New()
	gateway, err := h.gatewaySvc.GetB2CGateways(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get gateway: %v", err)
	}

	return &gatewaypb.GetB2CGatewaysResponse{
		Gateway: &gatewaypb.Gateway{
			Id:   gateway.ID,
			Name: gateway.Name,
		},
		Code: 200,
	}, nil
}

func (h *Handler) UpdateB2CGateways(ctx context.Context, req *gatewaypb.UpdateB2CGatewaysRequest) (*gatewaypb.UpdateB2CGatewaysResponse, error) {
	userID := uuid.New()
	err := h.gatewaySvc.SaveB2CGateways(userID, req.GatewayId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update gateway: %v", err)
	}

	return &gatewaypb.UpdateB2CGatewaysResponse{
		Message: "Gateway updated successfully",
		Code:    200,
	}, nil
}

func (h *Handler) DeleteB2CGateways(ctx context.Context, req *gatewaypb.DeleteB2CGatewaysRequest) (*gatewaypb.DeleteB2CGatewaysResponse, error) {
	userID := uuid.New()
	err := h.gatewaySvc.DeleteB2CGateways(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete gateway: %v", err)
	}

	return &gatewaypb.DeleteB2CGatewaysResponse{
		Message: "Gateway deleted successfully",
		Code:    200,
	}, nil
}

func (h *Handler) GetBaseAmount(ctx context.Context, req *baseamountpb.GetBaseAmountRequest) (*baseamountpb.GetBaseAmountResponse, error) {
	userID := ""
	baseAmounts, err := h.baseAmountSvc.GetBaseAmount(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get base amount: %v", err)
	}

	return &baseamountpb.GetBaseAmountResponse{
		UserId:              userID,
		MobileBaseAmount:    baseAmounts.MobileBaseAmount.Amount,
		UtilitiesBaseAmount: baseAmounts.UtilitiesBaseAmount.Amount,
		LeasingBaseAmount:   baseAmounts.LeasingBaseAmount.Amount,
		Code:                200,
	}, nil
}

func (h *Handler) CalculateMobileAmount(ctx context.Context, req *baseamountpb.CalculateMobileAmountRequest) (*baseamountpb.CalculateMobileAmountResponse, error) {
	userID := ""
	amount, err := h.baseAmountSvc.CalculateMobileAmount(userID, req.IsFirstMonth, req.Month)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate mobile amount: %v", err)
	}

	return &baseamountpb.CalculateMobileAmountResponse{
		UserId:       userID,
		Amount:       amount,
		IsFirstMonth: req.IsFirstMonth,
		Code:         200,
	}, nil
}

func (h *Handler) CalculateUtilitiesAmount(ctx context.Context, req *baseamountpb.CalculateUtilitiesAmountRequest) (*baseamountpb.CalculateUtilitiesAmountResponse, error) {
	userID := ""
	amount, err := h.baseAmountSvc.CalculateUtilitiesAmount(userID, req.IsFirstMonth, req.Month)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate utilities amount: %v", err)
	}

	return &baseamountpb.CalculateUtilitiesAmountResponse{
		UserId:       userID,
		Amount:       amount,
		IsFirstMonth: req.IsFirstMonth,
		Code:         200,
	}, nil
}

func (h *Handler) CalculateLeasingAmount(ctx context.Context, req *baseamountpb.CalculateLeasingAmountRequest) (*baseamountpb.CalculateLeasingAmountResponse, error) {
	userID := ""
	amount, err := h.baseAmountSvc.CalculateLeasingAmount(userID, req.Turnover, req.IsFirstMonth, req.Month)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate leasing amount: %v", err)
	}

	return &baseamountpb.CalculateLeasingAmountResponse{
		UserId:       userID,
		Amount:       amount,
		IsFirstMonth: req.IsFirstMonth,
		Turnover:     req.Turnover,
		Code:         200,
	}, nil
}

func (h *Handler) ResetMobileBaseAmount(ctx context.Context, req *baseamountpb.ResetMobileBaseAmountRequest) (*baseamountpb.ResetMobileBaseAmountResponse, error) {
	userID := ""
	err := h.baseAmountSvc.DeleteMobileBaseAmount(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset mobile base amount: %v", err)
	}

	return &baseamountpb.ResetMobileBaseAmountResponse{
		Message: "Mobile base amount reset successfully",
		Code:    200,
	}, nil
}

func (h *Handler) ResetUtilitiesBaseAmount(ctx context.Context, req *baseamountpb.ResetUtilitiesBaseAmountRequest) (*baseamountpb.ResetUtilitiesBaseAmountResponse, error) {
	userID := ""
	err := h.baseAmountSvc.DeleteUtilitiesBaseAmount(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset utilities base amount: %v", err)
	}

	return &baseamountpb.ResetUtilitiesBaseAmountResponse{
		Message: "Utilities base amount reset successfully",
		Code:    200,
	}, nil
}

func (h *Handler) ResetLeasingBaseAmount(ctx context.Context, req *baseamountpb.ResetLeasingBaseAmountRequest) (*baseamountpb.ResetLeasingBaseAmountResponse, error) {
	userID := ""
	err := h.baseAmountSvc.DeleteLeasingBaseAmount(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reset leasing base amount: %v", err)
	}

	return &baseamountpb.ResetLeasingBaseAmountResponse{
		Message: "Leasing base amount reset successfully",
		Code:    200,
	}, nil
}

func (h *Handler) CalculateRevenueBreakdown(ctx context.Context, req *breakdownpb.CalculateRevenueBreakdownRequest) (*breakdownpb.CalculateRevenueBreakdownResponse, error) {
	breakdown, err := h.breakdownSvc.GetRevenueBreakdown(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate revenue breakdown: %v", err)
	}

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
	breakdown, err := h.breakdownSvc.GetExpensesBreakdown(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate expenses breakdown: %v", err)
	}

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
	result, err := h.balanceAdjustmentSvc.ValidateBalance(req.RequestId)
	if err != nil {
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

	return &balancepb.ValidateBalanceResponse{
		RequestId: req.RequestId,
		IsValid:   result.IsValid,
		Issues:    issues,
		Code:      200,
	}, nil
}

func (h *Handler) GetBalanceAdjustment(ctx context.Context, req *balancepb.GetBalanceAdjustmentRequest) (*balancepb.GetBalanceAdjustmentResponse, error) {
	transactions, err := h.balanceAdjustmentSvc.GetAdjustedTransactions(req.RequestId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get balance adjustment: %v", err)
	}

	protoTxs := make([]*transactionpb.GeneratedTransaction, 0, len(transactions))
	for i := range transactions {
		protoTx := domainToProtoTransaction(&transactions[i], req.RequestId)
		protoTxs = append(protoTxs, protoTx)
	}

	return &balancepb.GetBalanceAdjustmentResponse{
		RequestId:    req.RequestId,
		Transactions: protoTxs,
		Code:         200,
	}, nil
}

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
