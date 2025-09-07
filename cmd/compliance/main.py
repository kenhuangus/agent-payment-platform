from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional

app = FastAPI()

class Party(BaseModel):
    id: str
    name: str
    type: str  # individual, organization
    created_at: str

class OnboardingStatus(BaseModel):
    party_id: str
    status: str
    details: Optional[str] = None

onboarding_db = {}

@app.get('/healthz')
def healthz():
    return {"status": "compliance service ok"}

@app.post('/onboard', response_model=OnboardingStatus)
def start_onboarding(party: Party):
    status = OnboardingStatus(party_id=party.id, status="pending", details="KYB/KYC started")
    onboarding_db[party.id] = status
    return status

@app.get('/onboard/{party_id}', response_model=OnboardingStatus)
def get_onboarding_status(party_id: str):
    status = onboarding_db.get(party_id)
    if not status:
        raise HTTPException(status_code=404, detail="Onboarding not found")
    return status
