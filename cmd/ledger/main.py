from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from typing import List, Optional

app = FastAPI()

class LedgerLine(BaseModel):
    account: str
    debit: Optional[float] = None
    credit: Optional[float] = None

class LedgerEntry(BaseModel):
    entry_id: str
    txn_id: str
    timestamp: str
    currency: str
    lines: List[LedgerLine]
    metadata: Optional[dict] = None
    hash: Optional[str] = None
    prev_hash: Optional[str] = None

ledger_db = {}

@app.get('/healthz')
def healthz():
    return {"status": "ledger service ok"}

@app.post('/ledger', response_model=LedgerEntry)
def post_ledger_entry(entry: LedgerEntry):
    if entry.entry_id in ledger_db:
        raise HTTPException(status_code=409, detail="Entry already exists")
    ledger_db[entry.entry_id] = entry
    return entry

@app.get('/ledger/{entry_id}', response_model=LedgerEntry)
def get_ledger_entry(entry_id: str):
    entry = ledger_db.get(entry_id)
    if not entry:
        raise HTTPException(status_code=404, detail="Entry not found")
    return entry
