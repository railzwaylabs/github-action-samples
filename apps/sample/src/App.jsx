import { useCallback, useEffect, useMemo, useState } from "react";

const initialForm = { customerId: "", productId: "", quantity: 1, price: 0 };
const apiBaseUrl = (import.meta.env.VITE_API_BASE_URL || "").replace(/\/$/, "");

function money(value) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" }).format(value / 100);
}

async function api(path, options) {
  const response = await fetch(`${apiBaseUrl}${path}`, options);
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error || `Request failed (${response.status})`);
  return body;
}

export default function App() {
  const [orders, setOrders] = useState([]);
  const [form, setForm] = useState(initialForm);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const loadOrders = useCallback(async () => {
    try {
      setError("");
      setOrders(await api("/orders"));
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { loadOrders(); }, [loadOrders]);

  const total = useMemo(
    () => orders.reduce((sum, order) => sum + order.items.reduce((itemSum, item) => itemSum + item.subtotal, 0), 0),
    [orders],
  );

  async function submit(event) {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      await api("/orders", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          customer_id: form.customerId,
          items: [{ product_id: form.productId, quantity: Number(form.quantity), price: Number(form.price) }],
        }),
      });
      setForm(initialForm);
      await loadOrders();
    } catch (requestError) {
      setError(requestError.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <main>
      <header>
        <div><h1>Order console</h1><p>Create and review persisted orders.</p></div>
        <span className="status">API connected</span>
      </header>

      <section className="summary">
        <span>Total booked</span><strong data-testid="total-booked">{money(total)}</strong><small>{orders.length} orders</small>
      </section>

      {error && <div className="alert" role="alert">{error}</div>}

      <div className="layout">
        <section className="panel">
          <div className="panel-heading"><h2>Create order</h2></div>
          <form onSubmit={submit}>
            <label>Customer ID<input data-testid="customer-id" required value={form.customerId} onChange={(e) => setForm({ ...form, customerId: e.target.value })} placeholder="customer-acme" /></label>
            <label>Product ID<input data-testid="product-id" required value={form.productId} onChange={(e) => setForm({ ...form, productId: e.target.value })} placeholder="api-growth" /></label>
            <div className="row">
              <label>Quantity<input data-testid="quantity" required min="1" type="number" value={form.quantity} onChange={(e) => setForm({ ...form, quantity: e.target.value })} /></label>
              <label>Unit price (cents)<input data-testid="price" required min="0" type="number" value={form.price} onChange={(e) => setForm({ ...form, price: e.target.value })} /></label>
            </div>
            <button data-testid="create-order" disabled={submitting}>{submitting ? "Creating…" : "Create order"}</button>
          </form>
        </section>

        <section className="panel orders">
          <div className="panel-heading"><h2>Recent orders</h2><button className="secondary" onClick={loadOrders}>Refresh</button></div>
          {loading ? <p className="empty">Loading orders…</p> : orders.length === 0 ? <p className="empty">No orders yet. Create the first one.</p> : (
            <div className="order-list" data-testid="order-list">
              {orders.map((order) => <article className="order" key={order.id} data-testid="order-row"><div><strong>{order.customer_id}</strong><small>{order.id.slice(0, 12)}</small></div><div><strong>{money(order.items.reduce((sum, item) => sum + item.subtotal, 0))}</strong><small>{order.items.length} item</small></div></article>)}
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
