interface JsonRpcResponse<T> {
  jsonrpc: "2.0";
  id: string | number;
  result?: T;
  error?: {
    code: number;
    message: string;
  };
}

interface RpcCommand {
  jsonrpc: "2.0";
  id: string | number;
  method: string;
  params?: any[];
}

export class RpcClientService {
  // return codes for errors
  public static ErrRedirectToLogin = -32302;
  public static ErrNoConnectionToOrganization = -32097;

  public async execute<T>(method: string, params?: any[]) {
    const id = Date.now();
    const body = {
      jsonrpc: "2.0",
      id,
      method,
      params,
    } as RpcCommand;
    // const options = {
    //   withCredentials: true,
    //   headers: {
    //     "Content-Type": "application/json",
    //     Accept: "application/json",
    //   },
    // };
    return await fetch("/rpc?" + method, {
      method: "POST",
      body: JSON.stringify(body),
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
    })
      .then((resp) => resp.json())
      .then((jsonRpcResp) => {
        if (jsonRpcResp.jsonrpc !== "2.0") {
          throw Error("Not a JSON-RPC response");
        }
        if (jsonRpcResp.id !== id) {
          throw Error("Not a valid JSON-RPC response");
        }
        if (jsonRpcResp.error) {
          return Promise.reject(jsonRpcResp.error);
        }
        return jsonRpcResp.result as T;
      });
  }
}
