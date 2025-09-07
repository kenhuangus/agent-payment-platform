from pydantic import BaseModel
from typing import List, Optional

class Party(BaseModel):
    id: str
    name: str
    type: str  # individual, organization
    created_at: str

class Agent(BaseModel):
    id: str
    display_name: str
    owner_party_id: str
    identity_mode: str  # did, oauth
    created_at: str

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

# Add Account, Counterparty, Transaction, Posting, Decision, Case as needed
