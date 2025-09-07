from fastapi import FastAPI
from pydantic import BaseModel
from typing import Optional

app = FastAPI()

class RouteRequest(BaseModel):
    agent_id: str
    amount: float
    currency: str
    counterparty: str
    consent_id: str
    memo: Optional[str] = None

class RouteResponse(BaseModel):
    status: str
    rail: str
    plan: str

@app.get('/healthz')
def healthz():
    return {"status": "router service ok"}

@app.post('/route', response_model=RouteResponse)
def route_payment(req: RouteRequest):
    # Simulate routing logic
    rail = "ach" if req.amount < 5000 else "wire"
    plan = "authorize, place hold, submit, settle, release hold, reconcile"
    return RouteResponse(status="routed", rail=rail, plan=plan)
