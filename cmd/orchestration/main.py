from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import Optional

app = FastAPI()

class OrchestrationRequest(BaseModel):
    agent_id: str
    consent_id: str
    amount: float
    currency: str
    counterparty: str
    memo: Optional[str] = None

class OrchestrationStatus(BaseModel):
    workflow_id: str
    status: str
    details: Optional[str] = None

orchestration_db = {}

@app.get('/healthz')
def healthz():
    return {"status": "orchestration service ok"}

@app.post('/orchestrate', response_model=OrchestrationStatus)
def start_orchestration(req: OrchestrationRequest):
    workflow_id = f"wf_{req.agent_id}_{req.amount}"
    status = OrchestrationStatus(workflow_id=workflow_id, status="started", details="Orchestration initiated")
    orchestration_db[workflow_id] = status
    return status

@app.get('/orchestrate/{workflow_id}', response_model=OrchestrationStatus)
def get_orchestration_status(workflow_id: str):
    status = orchestration_db.get(workflow_id)
    if not status:
        raise HTTPException(status_code=404, detail="Workflow not found")
    return status
