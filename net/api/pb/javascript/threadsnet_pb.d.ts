// package: threads.net.pb
// file: threadsnet.proto


import * as jspb from "google-protobuf";

export class GetHostIDRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetHostIDRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetHostIDRequest): GetHostIDRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetHostIDRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetHostIDRequest;
  static deserializeBinaryFromReader(message: GetHostIDRequest, reader: jspb.BinaryReader): GetHostIDRequest;
}

export namespace GetHostIDRequest {
  export type AsObject = {
  }
}

export class GetHostIDReply extends jspb.Message {
  getPeerid(): Uint8Array | string;
  getPeerid_asU8(): Uint8Array;
  getPeerid_asB64(): string;
  setPeerid(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetHostIDReply.AsObject;
  static toObject(includeInstance: boolean, msg: GetHostIDReply): GetHostIDReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetHostIDReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetHostIDReply;
  static deserializeBinaryFromReader(message: GetHostIDReply, reader: jspb.BinaryReader): GetHostIDReply;
}

export namespace GetHostIDReply {
  export type AsObject = {
    peerid: Uint8Array | string,
  }
}

export class GetTokenRequest extends jspb.Message {
  hasKey(): boolean;
  clearKey(): void;
  getKey(): string;
  setKey(value: string): void;

  hasSignature(): boolean;
  clearSignature(): void;
  getSignature(): Uint8Array | string;
  getSignature_asU8(): Uint8Array;
  getSignature_asB64(): string;
  setSignature(value: Uint8Array | string): void;

  getPayloadCase(): GetTokenRequest.PayloadCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetTokenRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetTokenRequest): GetTokenRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetTokenRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetTokenRequest;
  static deserializeBinaryFromReader(message: GetTokenRequest, reader: jspb.BinaryReader): GetTokenRequest;
}

export namespace GetTokenRequest {
  export type AsObject = {
    key: string,
    signature: Uint8Array | string,
  }

  export enum PayloadCase {
    PAYLOAD_NOT_SET = 0,
    KEY = 1,
    SIGNATURE = 2,
  }
}

export class GetTokenReply extends jspb.Message {
  hasChallenge(): boolean;
  clearChallenge(): void;
  getChallenge(): Uint8Array | string;
  getChallenge_asU8(): Uint8Array;
  getChallenge_asB64(): string;
  setChallenge(value: Uint8Array | string): void;

  hasToken(): boolean;
  clearToken(): void;
  getToken(): string;
  setToken(value: string): void;

  getPayloadCase(): GetTokenReply.PayloadCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetTokenReply.AsObject;
  static toObject(includeInstance: boolean, msg: GetTokenReply): GetTokenReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetTokenReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetTokenReply;
  static deserializeBinaryFromReader(message: GetTokenReply, reader: jspb.BinaryReader): GetTokenReply;
}

export namespace GetTokenReply {
  export type AsObject = {
    challenge: Uint8Array | string,
    token: string,
  }

  export enum PayloadCase {
    PAYLOAD_NOT_SET = 0,
    CHALLENGE = 1,
    TOKEN = 2,
  }
}

export class CreateThreadRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  hasKeys(): boolean;
  clearKeys(): void;
  getKeys(): Keys | undefined;
  setKeys(value?: Keys): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateThreadRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateThreadRequest): CreateThreadRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateThreadRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateThreadRequest;
  static deserializeBinaryFromReader(message: CreateThreadRequest, reader: jspb.BinaryReader): CreateThreadRequest;
}

export namespace CreateThreadRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
    keys?: Keys.AsObject,
  }
}

export class Keys extends jspb.Message {
  getThreadkey(): Uint8Array | string;
  getThreadkey_asU8(): Uint8Array;
  getThreadkey_asB64(): string;
  setThreadkey(value: Uint8Array | string): void;

  getLogkey(): Uint8Array | string;
  getLogkey_asU8(): Uint8Array;
  getLogkey_asB64(): string;
  setLogkey(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Keys.AsObject;
  static toObject(includeInstance: boolean, msg: Keys): Keys.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Keys, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Keys;
  static deserializeBinaryFromReader(message: Keys, reader: jspb.BinaryReader): Keys;
}

export namespace Keys {
  export type AsObject = {
    threadkey: Uint8Array | string,
    logkey: Uint8Array | string,
  }
}

export class ThreadInfoReply extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  getThreadkey(): Uint8Array | string;
  getThreadkey_asU8(): Uint8Array;
  getThreadkey_asB64(): string;
  setThreadkey(value: Uint8Array | string): void;

  clearLogsList(): void;
  getLogsList(): Array<LogInfo>;
  setLogsList(value: Array<LogInfo>): void;
  addLogs(value?: LogInfo, index?: number): LogInfo;

  clearAddrsList(): void;
  getAddrsList(): Array<Uint8Array | string>;
  getAddrsList_asU8(): Array<Uint8Array>;
  getAddrsList_asB64(): Array<string>;
  setAddrsList(value: Array<Uint8Array | string>): void;
  addAddrs(value: Uint8Array | string, index?: number): Uint8Array | string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ThreadInfoReply.AsObject;
  static toObject(includeInstance: boolean, msg: ThreadInfoReply): ThreadInfoReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ThreadInfoReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ThreadInfoReply;
  static deserializeBinaryFromReader(message: ThreadInfoReply, reader: jspb.BinaryReader): ThreadInfoReply;
}

export namespace ThreadInfoReply {
  export type AsObject = {
    threadid: Uint8Array | string,
    threadkey: Uint8Array | string,
    logsList: Array<LogInfo.AsObject>,
    addrsList: Array<Uint8Array | string>,
  }
}

export class LogInfo extends jspb.Message {
  getId(): Uint8Array | string;
  getId_asU8(): Uint8Array;
  getId_asB64(): string;
  setId(value: Uint8Array | string): void;

  getPubkey(): Uint8Array | string;
  getPubkey_asU8(): Uint8Array;
  getPubkey_asB64(): string;
  setPubkey(value: Uint8Array | string): void;

  getPrivkey(): Uint8Array | string;
  getPrivkey_asU8(): Uint8Array;
  getPrivkey_asB64(): string;
  setPrivkey(value: Uint8Array | string): void;

  clearAddrsList(): void;
  getAddrsList(): Array<Uint8Array | string>;
  getAddrsList_asU8(): Array<Uint8Array>;
  getAddrsList_asB64(): Array<string>;
  setAddrsList(value: Array<Uint8Array | string>): void;
  addAddrs(value: Uint8Array | string, index?: number): Uint8Array | string;

  getHead(): Uint8Array | string;
  getHead_asU8(): Uint8Array;
  getHead_asB64(): string;
  setHead(value: Uint8Array | string): void;

  getCounter(): Uint8Array | string;
  getCounter_asU8(): Uint8Array;
  getCounter_asB64(): string;
  setCounter(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): LogInfo.AsObject;
  static toObject(includeInstance: boolean, msg: LogInfo): LogInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: LogInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): LogInfo;
  static deserializeBinaryFromReader(message: LogInfo, reader: jspb.BinaryReader): LogInfo;
}

export namespace LogInfo {
  export type AsObject = {
    id: Uint8Array | string,
    pubkey: Uint8Array | string,
    privkey: Uint8Array | string,
    addrsList: Array<Uint8Array | string>,
    head: Uint8Array | string,
    counter: Uint8Array | string,
  }
}

export class AddThreadRequest extends jspb.Message {
  getAddr(): Uint8Array | string;
  getAddr_asU8(): Uint8Array;
  getAddr_asB64(): string;
  setAddr(value: Uint8Array | string): void;

  hasKeys(): boolean;
  clearKeys(): void;
  getKeys(): Keys | undefined;
  setKeys(value?: Keys): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddThreadRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddThreadRequest): AddThreadRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddThreadRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddThreadRequest;
  static deserializeBinaryFromReader(message: AddThreadRequest, reader: jspb.BinaryReader): AddThreadRequest;
}

export namespace AddThreadRequest {
  export type AsObject = {
    addr: Uint8Array | string,
    keys?: Keys.AsObject,
  }
}

export class GetThreadRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetThreadRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetThreadRequest): GetThreadRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetThreadRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetThreadRequest;
  static deserializeBinaryFromReader(message: GetThreadRequest, reader: jspb.BinaryReader): GetThreadRequest;
}

export namespace GetThreadRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
  }
}

export class PullThreadRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PullThreadRequest.AsObject;
  static toObject(includeInstance: boolean, msg: PullThreadRequest): PullThreadRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PullThreadRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PullThreadRequest;
  static deserializeBinaryFromReader(message: PullThreadRequest, reader: jspb.BinaryReader): PullThreadRequest;
}

export namespace PullThreadRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
  }
}

export class PullThreadReply extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): PullThreadReply.AsObject;
  static toObject(includeInstance: boolean, msg: PullThreadReply): PullThreadReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: PullThreadReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): PullThreadReply;
  static deserializeBinaryFromReader(message: PullThreadReply, reader: jspb.BinaryReader): PullThreadReply;
}

export namespace PullThreadReply {
  export type AsObject = {
  }
}

export class DeleteThreadRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteThreadRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteThreadRequest): DeleteThreadRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteThreadRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteThreadRequest;
  static deserializeBinaryFromReader(message: DeleteThreadRequest, reader: jspb.BinaryReader): DeleteThreadRequest;
}

export namespace DeleteThreadRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
  }
}

export class DeleteThreadReply extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteThreadReply.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteThreadReply): DeleteThreadReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteThreadReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteThreadReply;
  static deserializeBinaryFromReader(message: DeleteThreadReply, reader: jspb.BinaryReader): DeleteThreadReply;
}

export namespace DeleteThreadReply {
  export type AsObject = {
  }
}

export class AddReplicatorRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  getAddr(): Uint8Array | string;
  getAddr_asU8(): Uint8Array;
  getAddr_asB64(): string;
  setAddr(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddReplicatorRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddReplicatorRequest): AddReplicatorRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddReplicatorRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddReplicatorRequest;
  static deserializeBinaryFromReader(message: AddReplicatorRequest, reader: jspb.BinaryReader): AddReplicatorRequest;
}

export namespace AddReplicatorRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
    addr: Uint8Array | string,
  }
}

export class AddReplicatorReply extends jspb.Message {
  getPeerid(): Uint8Array | string;
  getPeerid_asU8(): Uint8Array;
  getPeerid_asB64(): string;
  setPeerid(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddReplicatorReply.AsObject;
  static toObject(includeInstance: boolean, msg: AddReplicatorReply): AddReplicatorReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddReplicatorReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddReplicatorReply;
  static deserializeBinaryFromReader(message: AddReplicatorReply, reader: jspb.BinaryReader): AddReplicatorReply;
}

export namespace AddReplicatorReply {
  export type AsObject = {
    peerid: Uint8Array | string,
  }
}

export class CreateRecordRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  getBody(): Uint8Array | string;
  getBody_asU8(): Uint8Array;
  getBody_asB64(): string;
  setBody(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateRecordRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateRecordRequest): CreateRecordRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateRecordRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateRecordRequest;
  static deserializeBinaryFromReader(message: CreateRecordRequest, reader: jspb.BinaryReader): CreateRecordRequest;
}

export namespace CreateRecordRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
    body: Uint8Array | string,
  }
}

export class NewRecordReply extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  getLogid(): Uint8Array | string;
  getLogid_asU8(): Uint8Array;
  getLogid_asB64(): string;
  setLogid(value: Uint8Array | string): void;

  hasRecord(): boolean;
  clearRecord(): void;
  getRecord(): Record | undefined;
  setRecord(value?: Record): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): NewRecordReply.AsObject;
  static toObject(includeInstance: boolean, msg: NewRecordReply): NewRecordReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: NewRecordReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): NewRecordReply;
  static deserializeBinaryFromReader(message: NewRecordReply, reader: jspb.BinaryReader): NewRecordReply;
}

export namespace NewRecordReply {
  export type AsObject = {
    threadid: Uint8Array | string,
    logid: Uint8Array | string,
    record?: Record.AsObject,
  }
}

export class AddRecordRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  getLogid(): Uint8Array | string;
  getLogid_asU8(): Uint8Array;
  getLogid_asB64(): string;
  setLogid(value: Uint8Array | string): void;

  hasRecord(): boolean;
  clearRecord(): void;
  getRecord(): Record | undefined;
  setRecord(value?: Record): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddRecordRequest.AsObject;
  static toObject(includeInstance: boolean, msg: AddRecordRequest): AddRecordRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddRecordRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddRecordRequest;
  static deserializeBinaryFromReader(message: AddRecordRequest, reader: jspb.BinaryReader): AddRecordRequest;
}

export namespace AddRecordRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
    logid: Uint8Array | string,
    record?: Record.AsObject,
  }
}

export class Record extends jspb.Message {
  getRecordnode(): Uint8Array | string;
  getRecordnode_asU8(): Uint8Array;
  getRecordnode_asB64(): string;
  setRecordnode(value: Uint8Array | string): void;

  getEventnode(): Uint8Array | string;
  getEventnode_asU8(): Uint8Array;
  getEventnode_asB64(): string;
  setEventnode(value: Uint8Array | string): void;

  getHeadernode(): Uint8Array | string;
  getHeadernode_asU8(): Uint8Array;
  getHeadernode_asB64(): string;
  setHeadernode(value: Uint8Array | string): void;

  getBodynode(): Uint8Array | string;
  getBodynode_asU8(): Uint8Array;
  getBodynode_asB64(): string;
  setBodynode(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Record.AsObject;
  static toObject(includeInstance: boolean, msg: Record): Record.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Record, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Record;
  static deserializeBinaryFromReader(message: Record, reader: jspb.BinaryReader): Record;
}

export namespace Record {
  export type AsObject = {
    recordnode: Uint8Array | string,
    eventnode: Uint8Array | string,
    headernode: Uint8Array | string,
    bodynode: Uint8Array | string,
  }
}

export class AddRecordReply extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AddRecordReply.AsObject;
  static toObject(includeInstance: boolean, msg: AddRecordReply): AddRecordReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AddRecordReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AddRecordReply;
  static deserializeBinaryFromReader(message: AddRecordReply, reader: jspb.BinaryReader): AddRecordReply;
}

export namespace AddRecordReply {
  export type AsObject = {
  }
}

export class GetRecordRequest extends jspb.Message {
  getThreadid(): Uint8Array | string;
  getThreadid_asU8(): Uint8Array;
  getThreadid_asB64(): string;
  setThreadid(value: Uint8Array | string): void;

  getRecordid(): Uint8Array | string;
  getRecordid_asU8(): Uint8Array;
  getRecordid_asB64(): string;
  setRecordid(value: Uint8Array | string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetRecordRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetRecordRequest): GetRecordRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetRecordRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetRecordRequest;
  static deserializeBinaryFromReader(message: GetRecordRequest, reader: jspb.BinaryReader): GetRecordRequest;
}

export namespace GetRecordRequest {
  export type AsObject = {
    threadid: Uint8Array | string,
    recordid: Uint8Array | string,
  }
}

export class GetRecordReply extends jspb.Message {
  hasRecord(): boolean;
  clearRecord(): void;
  getRecord(): Record | undefined;
  setRecord(value?: Record): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetRecordReply.AsObject;
  static toObject(includeInstance: boolean, msg: GetRecordReply): GetRecordReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetRecordReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetRecordReply;
  static deserializeBinaryFromReader(message: GetRecordReply, reader: jspb.BinaryReader): GetRecordReply;
}

export namespace GetRecordReply {
  export type AsObject = {
    record?: Record.AsObject,
  }
}

export class SubscribeRequest extends jspb.Message {
  clearThreadidsList(): void;
  getThreadidsList(): Array<Uint8Array | string>;
  getThreadidsList_asU8(): Array<Uint8Array>;
  getThreadidsList_asB64(): Array<string>;
  setThreadidsList(value: Array<Uint8Array | string>): void;
  addThreadids(value: Uint8Array | string, index?: number): Uint8Array | string;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubscribeRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubscribeRequest): SubscribeRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SubscribeRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubscribeRequest;
  static deserializeBinaryFromReader(message: SubscribeRequest, reader: jspb.BinaryReader): SubscribeRequest;
}

export namespace SubscribeRequest {
  export type AsObject = {
    threadidsList: Array<Uint8Array | string>,
  }
}

