from typing import Tuple

def getHeadersInfo(headers) -> Tuple[str, str]:
    host, access_token = str(headers.get('x-as-address')), str(headers.get('x-authorization'))
    if access_token == '' or access_token == 'None':
        access_token = str(headers.get('Authorization'))
    if access_token.startswith("Bearer "):
        access_token = access_token[len("Bearer "):]
    return host, access_token