// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.0
// source: shared/protocol/account/account.proto

package account

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CurrencyCreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Currency    string                 `protobuf:"bytes,1,opt,name=currency,proto3" json:"currency,omitempty"`
	CreditLimit *wrapperspb.Int64Value `protobuf:"bytes,2,opt,name=creditLimit,proto3" json:"creditLimit,omitempty"`
}

func (x *CurrencyCreateRequest) Reset() {
	*x = CurrencyCreateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_protocol_account_account_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CurrencyCreateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CurrencyCreateRequest) ProtoMessage() {}

func (x *CurrencyCreateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_shared_protocol_account_account_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CurrencyCreateRequest.ProtoReflect.Descriptor instead.
func (*CurrencyCreateRequest) Descriptor() ([]byte, []int) {
	return file_shared_protocol_account_account_proto_rawDescGZIP(), []int{0}
}

func (x *CurrencyCreateRequest) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *CurrencyCreateRequest) GetCreditLimit() *wrapperspb.Int64Value {
	if x != nil {
		return x.CreditLimit
	}
	return nil
}

type CreateAccountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AllowAsync bool                     `protobuf:"varint,1,opt,name=allowAsync,proto3" json:"allowAsync,omitempty"`
	Currencies []*CurrencyCreateRequest `protobuf:"bytes,2,rep,name=currencies,proto3" json:"currencies,omitempty"`
}

func (x *CreateAccountRequest) Reset() {
	*x = CreateAccountRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_protocol_account_account_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateAccountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateAccountRequest) ProtoMessage() {}

func (x *CreateAccountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_shared_protocol_account_account_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateAccountRequest.ProtoReflect.Descriptor instead.
func (*CreateAccountRequest) Descriptor() ([]byte, []int) {
	return file_shared_protocol_account_account_proto_rawDescGZIP(), []int{1}
}

func (x *CreateAccountRequest) GetAllowAsync() bool {
	if x != nil {
		return x.AllowAsync
	}
	return false
}

func (x *CreateAccountRequest) GetCurrencies() []*CurrencyCreateRequest {
	if x != nil {
		return x.Currencies
	}
	return nil
}

type AccountCurrencyDetailsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Currency    string                 `protobuf:"bytes,1,opt,name=currency,proto3" json:"currency,omitempty"`
	Amount      int64                  `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
	CreditLimit *wrapperspb.Int64Value `protobuf:"bytes,3,opt,name=creditLimit,proto3" json:"creditLimit,omitempty"`
}

func (x *AccountCurrencyDetailsResponse) Reset() {
	*x = AccountCurrencyDetailsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_protocol_account_account_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountCurrencyDetailsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountCurrencyDetailsResponse) ProtoMessage() {}

func (x *AccountCurrencyDetailsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_shared_protocol_account_account_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountCurrencyDetailsResponse.ProtoReflect.Descriptor instead.
func (*AccountCurrencyDetailsResponse) Descriptor() ([]byte, []int) {
	return file_shared_protocol_account_account_proto_rawDescGZIP(), []int{2}
}

func (x *AccountCurrencyDetailsResponse) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *AccountCurrencyDetailsResponse) GetAmount() int64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *AccountCurrencyDetailsResponse) GetCreditLimit() *wrapperspb.Int64Value {
	if x != nil {
		return x.CreditLimit
	}
	return nil
}

type AccountDetailsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id              []byte                            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	IsSuspended     bool                              `protobuf:"varint,2,opt,name=isSuspended,proto3" json:"isSuspended,omitempty"`
	AllowAsync      bool                              `protobuf:"varint,3,opt,name=allowAsync,proto3" json:"allowAsync,omitempty"`
	SuspendedReason *wrapperspb.StringValue           `protobuf:"bytes,4,opt,name=suspendedReason,proto3" json:"suspendedReason,omitempty"`
	Currencies      []*AccountCurrencyDetailsResponse `protobuf:"bytes,5,rep,name=currencies,proto3" json:"currencies,omitempty"`
}

func (x *AccountDetailsResponse) Reset() {
	*x = AccountDetailsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_protocol_account_account_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountDetailsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountDetailsResponse) ProtoMessage() {}

func (x *AccountDetailsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_shared_protocol_account_account_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountDetailsResponse.ProtoReflect.Descriptor instead.
func (*AccountDetailsResponse) Descriptor() ([]byte, []int) {
	return file_shared_protocol_account_account_proto_rawDescGZIP(), []int{3}
}

func (x *AccountDetailsResponse) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *AccountDetailsResponse) GetIsSuspended() bool {
	if x != nil {
		return x.IsSuspended
	}
	return false
}

func (x *AccountDetailsResponse) GetAllowAsync() bool {
	if x != nil {
		return x.AllowAsync
	}
	return false
}

func (x *AccountDetailsResponse) GetSuspendedReason() *wrapperspb.StringValue {
	if x != nil {
		return x.SuspendedReason
	}
	return nil
}

func (x *AccountDetailsResponse) GetCurrencies() []*AccountCurrencyDetailsResponse {
	if x != nil {
		return x.Currencies
	}
	return nil
}

type CreateAccountResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status  int32                   `protobuf:"varint,1,opt,name=status,proto3" json:"status,omitempty"`
	Details *AccountDetailsResponse `protobuf:"bytes,2,opt,name=details,proto3" json:"details,omitempty"`
}

func (x *CreateAccountResponse) Reset() {
	*x = CreateAccountResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_protocol_account_account_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateAccountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateAccountResponse) ProtoMessage() {}

func (x *CreateAccountResponse) ProtoReflect() protoreflect.Message {
	mi := &file_shared_protocol_account_account_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateAccountResponse.ProtoReflect.Descriptor instead.
func (*CreateAccountResponse) Descriptor() ([]byte, []int) {
	return file_shared_protocol_account_account_proto_rawDescGZIP(), []int{4}
}

func (x *CreateAccountResponse) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *CreateAccountResponse) GetDetails() *AccountDetailsResponse {
	if x != nil {
		return x.Details
	}
	return nil
}

type GetAccountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetAccountRequest) Reset() {
	*x = GetAccountRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_shared_protocol_account_account_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAccountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAccountRequest) ProtoMessage() {}

func (x *GetAccountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_shared_protocol_account_account_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAccountRequest.ProtoReflect.Descriptor instead.
func (*GetAccountRequest) Descriptor() ([]byte, []int) {
	return file_shared_protocol_account_account_proto_rawDescGZIP(), []int{5}
}

func (x *GetAccountRequest) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

var File_shared_protocol_account_account_proto protoreflect.FileDescriptor

var file_shared_protocol_account_account_proto_rawDesc = []byte{
	0x0a, 0x25, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f,
	0x6c, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x72, 0x0a, 0x15, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x63, 0x79, 0x12, 0x3d, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x69, 0x74, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74,
	0x36, 0x34, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x69, 0x74, 0x4c,
	0x69, 0x6d, 0x69, 0x74, 0x22, 0x76, 0x0a, 0x14, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a,
	0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x41, 0x73, 0x79, 0x6e, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x41, 0x73, 0x79, 0x6e, 0x63, 0x12, 0x3e, 0x0a, 0x0a,
	0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1e, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x43, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x63, 0x79, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x52, 0x0a, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0x22, 0x93, 0x01, 0x0a,
	0x1e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79,
	0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x61, 0x6d, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x3d, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x69, 0x74, 0x4c, 0x69, 0x6d,
	0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x49, 0x6e, 0x74, 0x36, 0x34,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x69, 0x74, 0x4c, 0x69, 0x6d,
	0x69, 0x74, 0x22, 0xfb, 0x01, 0x0a, 0x16, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65,
	0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x12, 0x20, 0x0a,
	0x0b, 0x69, 0x73, 0x53, 0x75, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0b, 0x69, 0x73, 0x53, 0x75, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x64, 0x12,
	0x1e, 0x0a, 0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x41, 0x73, 0x79, 0x6e, 0x63, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x0a, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x41, 0x73, 0x79, 0x6e, 0x63, 0x12,
	0x46, 0x0a, 0x0f, 0x73, 0x75, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x65, 0x64, 0x52, 0x65, 0x61, 0x73,
	0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x0f, 0x73, 0x75, 0x73, 0x70, 0x65, 0x6e, 0x64, 0x65,
	0x64, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x47, 0x0a, 0x0a, 0x63, 0x75, 0x72, 0x72, 0x65,
	0x6e, 0x63, 0x69, 0x65, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x61, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x43, 0x75, 0x72,
	0x72, 0x65, 0x6e, 0x63, 0x79, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x52, 0x0a, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73,
	0x22, 0x6a, 0x0a, 0x15, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x39, 0x0a, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x52, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x22, 0x23, 0x0a, 0x11,
	0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69,
	0x64, 0x32, 0xa8, 0x01, 0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x50, 0x0a,
	0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d,
	0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x4b, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1a, 0x2e,
	0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e, 0x47, 0x65, 0x74, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x14, 0x5a, 0x12,
	0x6c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_shared_protocol_account_account_proto_rawDescOnce sync.Once
	file_shared_protocol_account_account_proto_rawDescData = file_shared_protocol_account_account_proto_rawDesc
)

func file_shared_protocol_account_account_proto_rawDescGZIP() []byte {
	file_shared_protocol_account_account_proto_rawDescOnce.Do(func() {
		file_shared_protocol_account_account_proto_rawDescData = protoimpl.X.CompressGZIP(file_shared_protocol_account_account_proto_rawDescData)
	})
	return file_shared_protocol_account_account_proto_rawDescData
}

var file_shared_protocol_account_account_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_shared_protocol_account_account_proto_goTypes = []interface{}{
	(*CurrencyCreateRequest)(nil),          // 0: account.CurrencyCreateRequest
	(*CreateAccountRequest)(nil),           // 1: account.CreateAccountRequest
	(*AccountCurrencyDetailsResponse)(nil), // 2: account.AccountCurrencyDetailsResponse
	(*AccountDetailsResponse)(nil),         // 3: account.AccountDetailsResponse
	(*CreateAccountResponse)(nil),          // 4: account.CreateAccountResponse
	(*GetAccountRequest)(nil),              // 5: account.GetAccountRequest
	(*wrapperspb.Int64Value)(nil),          // 6: google.protobuf.Int64Value
	(*wrapperspb.StringValue)(nil),         // 7: google.protobuf.StringValue
}
var file_shared_protocol_account_account_proto_depIdxs = []int32{
	6, // 0: account.CurrencyCreateRequest.creditLimit:type_name -> google.protobuf.Int64Value
	0, // 1: account.CreateAccountRequest.currencies:type_name -> account.CurrencyCreateRequest
	6, // 2: account.AccountCurrencyDetailsResponse.creditLimit:type_name -> google.protobuf.Int64Value
	7, // 3: account.AccountDetailsResponse.suspendedReason:type_name -> google.protobuf.StringValue
	2, // 4: account.AccountDetailsResponse.currencies:type_name -> account.AccountCurrencyDetailsResponse
	3, // 5: account.CreateAccountResponse.details:type_name -> account.AccountDetailsResponse
	1, // 6: account.Account.CreateAccount:input_type -> account.CreateAccountRequest
	5, // 7: account.Account.GetAccount:input_type -> account.GetAccountRequest
	4, // 8: account.Account.CreateAccount:output_type -> account.CreateAccountResponse
	3, // 9: account.Account.GetAccount:output_type -> account.AccountDetailsResponse
	8, // [8:10] is the sub-list for method output_type
	6, // [6:8] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_shared_protocol_account_account_proto_init() }
func file_shared_protocol_account_account_proto_init() {
	if File_shared_protocol_account_account_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_shared_protocol_account_account_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CurrencyCreateRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_shared_protocol_account_account_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateAccountRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_shared_protocol_account_account_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountCurrencyDetailsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_shared_protocol_account_account_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountDetailsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_shared_protocol_account_account_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateAccountResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_shared_protocol_account_account_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAccountRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_shared_protocol_account_account_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_shared_protocol_account_account_proto_goTypes,
		DependencyIndexes: file_shared_protocol_account_account_proto_depIdxs,
		MessageInfos:      file_shared_protocol_account_account_proto_msgTypes,
	}.Build()
	File_shared_protocol_account_account_proto = out.File
	file_shared_protocol_account_account_proto_rawDesc = nil
	file_shared_protocol_account_account_proto_goTypes = nil
	file_shared_protocol_account_account_proto_depIdxs = nil
}
