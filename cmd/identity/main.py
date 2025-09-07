from fastapi import FastAPI
app = FastAPI()

@app.get('/healthz')
def healthz():
    return {"status": "identity service ok"}
