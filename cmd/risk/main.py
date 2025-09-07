from fastapi import FastAPI
from pydantic import BaseModel
from typing import Optional

app = FastAPI()

class RiskRequest(BaseModel):
    agent_id: str
    amount: float
    currency: str
    counterparty: str
    consent_id: str
    context: Optional[dict] = None

class RiskResponse(BaseModel):
    score: float
    decision: str
    reason: str

@app.get('/healthz')
def healthz():
    return {"status": "risk service ok"}

@app.post('/score', response_model=RiskResponse)
def score_risk(req: RiskRequest):
    # Simulate scoring logic
    score = 0.1 if req.amount < 1000 else 0.8
    decision = "approve" if score < 0.5 else "review"
    reason = "Low amount" if score < 0.5 else "High amount"
    return RiskResponse(score=score, decision=decision, reason=reason)
