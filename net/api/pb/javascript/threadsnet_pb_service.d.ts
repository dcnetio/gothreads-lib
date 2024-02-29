// package: threads.net.pb
// file: threadsnet.proto

import * as threadsnet_pb from "./threadsnet_pb";
import {grpc} from "@improbable-eng/grpc-web";


type APIGetHostID = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.GetHostIDRequest;
  readonly responseType: typeof threadsnet_pb.GetHostIDReply;
};

type APIGetToken = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: true;
  readonly responseStream: true;
  readonly requestType: typeof threadsnet_pb.GetTokenRequest;
  readonly responseType: typeof threadsnet_pb.GetTokenReply;
};

type APICreateThread = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.CreateThreadRequest;
  readonly responseType: typeof threadsnet_pb.ThreadInfoReply;
};

type APIAddThread = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.AddThreadRequest;
  readonly responseType: typeof threadsnet_pb.ThreadInfoReply;
};

type APIGetThread = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.GetThreadRequest;
  readonly responseType: typeof threadsnet_pb.ThreadInfoReply;
};

type APIPullThread = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.PullThreadRequest;
  readonly responseType: typeof threadsnet_pb.PullThreadReply;
};

type APIDeleteThread = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.DeleteThreadRequest;
  readonly responseType: typeof threadsnet_pb.DeleteThreadReply;
};

type APIAddReplicator = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.AddReplicatorRequest;
  readonly responseType: typeof threadsnet_pb.AddReplicatorReply;
};

type APICreateRecord = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.CreateRecordRequest;
  readonly responseType: typeof threadsnet_pb.NewRecordReply;
};

type APIAddRecord = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.AddRecordRequest;
  readonly responseType: typeof threadsnet_pb.AddRecordReply;
};

type APIGetRecord = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof threadsnet_pb.GetRecordRequest;
  readonly responseType: typeof threadsnet_pb.GetRecordReply;
};

type APISubscribe = {
  readonly methodName: string;
  readonly service: typeof API;
  readonly requestStream: false;
  readonly responseStream: true;
  readonly requestType: typeof threadsnet_pb.SubscribeRequest;
  readonly responseType: typeof threadsnet_pb.NewRecordReply;
};

export class API {
  static readonly serviceName: string;
  static readonly GetHostID: APIGetHostID;
  static readonly GetToken: APIGetToken;
  static readonly CreateThread: APICreateThread;
  static readonly AddThread: APIAddThread;
  static readonly GetThread: APIGetThread;
  static readonly PullThread: APIPullThread;
  static readonly DeleteThread: APIDeleteThread;
  static readonly AddReplicator: APIAddReplicator;
  static readonly CreateRecord: APICreateRecord;
  static readonly AddRecord: APIAddRecord;
  static readonly GetRecord: APIGetRecord;
  static readonly Subscribe: APISubscribe;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class APIClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  getHostID(
    requestMessage: threadsnet_pb.GetHostIDRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.GetHostIDReply|null) => void
  ): UnaryResponse;
  getHostID(
    requestMessage: threadsnet_pb.GetHostIDRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.GetHostIDReply|null) => void
  ): UnaryResponse;
  getToken(metadata?: grpc.Metadata): BidirectionalStream<threadsnet_pb.GetTokenRequest, threadsnet_pb.GetTokenReply>;
  createThread(
    requestMessage: threadsnet_pb.CreateThreadRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.ThreadInfoReply|null) => void
  ): UnaryResponse;
  createThread(
    requestMessage: threadsnet_pb.CreateThreadRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.ThreadInfoReply|null) => void
  ): UnaryResponse;
  addThread(
    requestMessage: threadsnet_pb.AddThreadRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.ThreadInfoReply|null) => void
  ): UnaryResponse;
  addThread(
    requestMessage: threadsnet_pb.AddThreadRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.ThreadInfoReply|null) => void
  ): UnaryResponse;
  getThread(
    requestMessage: threadsnet_pb.GetThreadRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.ThreadInfoReply|null) => void
  ): UnaryResponse;
  getThread(
    requestMessage: threadsnet_pb.GetThreadRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.ThreadInfoReply|null) => void
  ): UnaryResponse;
  pullThread(
    requestMessage: threadsnet_pb.PullThreadRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.PullThreadReply|null) => void
  ): UnaryResponse;
  pullThread(
    requestMessage: threadsnet_pb.PullThreadRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.PullThreadReply|null) => void
  ): UnaryResponse;
  deleteThread(
    requestMessage: threadsnet_pb.DeleteThreadRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.DeleteThreadReply|null) => void
  ): UnaryResponse;
  deleteThread(
    requestMessage: threadsnet_pb.DeleteThreadRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.DeleteThreadReply|null) => void
  ): UnaryResponse;
  addReplicator(
    requestMessage: threadsnet_pb.AddReplicatorRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.AddReplicatorReply|null) => void
  ): UnaryResponse;
  addReplicator(
    requestMessage: threadsnet_pb.AddReplicatorRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.AddReplicatorReply|null) => void
  ): UnaryResponse;
  createRecord(
    requestMessage: threadsnet_pb.CreateRecordRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.NewRecordReply|null) => void
  ): UnaryResponse;
  createRecord(
    requestMessage: threadsnet_pb.CreateRecordRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.NewRecordReply|null) => void
  ): UnaryResponse;
  addRecord(
    requestMessage: threadsnet_pb.AddRecordRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.AddRecordReply|null) => void
  ): UnaryResponse;
  addRecord(
    requestMessage: threadsnet_pb.AddRecordRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.AddRecordReply|null) => void
  ): UnaryResponse;
  getRecord(
    requestMessage: threadsnet_pb.GetRecordRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.GetRecordReply|null) => void
  ): UnaryResponse;
  getRecord(
    requestMessage: threadsnet_pb.GetRecordRequest,
    callback: (error: ServiceError|null, responseMessage: threadsnet_pb.GetRecordReply|null) => void
  ): UnaryResponse;
  subscribe(requestMessage: threadsnet_pb.SubscribeRequest, metadata?: grpc.Metadata): ResponseStream<threadsnet_pb.NewRecordReply>;
}

