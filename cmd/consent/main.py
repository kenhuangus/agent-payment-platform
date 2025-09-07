from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI()

class ConsentLimits(BaseModel):
    single_txn_usd: float
    daily_usd: float
    max_txn_per_hour: int

class CosignRule(BaseModel):
    threshold_usd: float
    approver_group: str

class Consent(BaseModel):
    id: str
    agent_id: str
    owner_party_id: str
    rails: List[str]
    counterparties_allow: List[str]
    limits: ConsentLimits
    policy_bundle_version: str
    cosign_rule: CosignRule
    created_at: str
    revoked: bool

consents_db = {}

@app.get('/healthz')
def healthz():
    return {"status": "consent service ok"}

@app.post('/consents', response_model=Consent)
def create_consent(consent: Consent):
    if consent.id in consents_db:
        raise HTTPException(status_code=409, detail="Consent already exists")
    consents_db[consent.id] = consent
    return consent

@app.get('/consents/{consent_id}', response_model=Consent)
def get_consent(consent_id: str):
    consent = consents_db.get(consent_id)
    if not consent:
        raise HTTPException(status_code=404, detail="Consent not found")
    return consent
