CREATE OR REPLACE FUNCTION public.set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_users_set_updated_at
BEFORE UPDATE ON public.users
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

CREATE TRIGGER trg_gems_set_updated_at
BEFORE UPDATE ON public.gems
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

CREATE TRIGGER trg_listings_set_updated_at
BEFORE UPDATE ON public.listings
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

CREATE TRIGGER trg_orders_set_updated_at
BEFORE UPDATE ON public.orders
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

CREATE TRIGGER trg_shipping_set_updated_at
BEFORE UPDATE ON public.shipping
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

CREATE INDEX idx_seller_documents_seller_id
    ON public.seller_documents(seller_id);

CREATE INDEX idx_gems_seller_id
    ON public.gems(seller_id);

CREATE INDEX idx_gems_gem_type
    ON public.gems(gem_type);

CREATE INDEX idx_gem_images_gem_id
    ON public.gem_images(gem_id);

CREATE UNIQUE INDEX uq_gem_images_one_primary_per_gem
    ON public.gem_images(gem_id)
    WHERE is_primary = TRUE;

CREATE INDEX idx_certificates_gem_id
    ON public.certificates(gem_id);

CREATE INDEX idx_listings_gem_id
    ON public.listings(gem_id);

CREATE INDEX idx_listings_seller_status
    ON public.listings(seller_id, status);

CREATE INDEX idx_listings_sale_type_status
    ON public.listings(sale_type, status);

CREATE INDEX idx_auctions_status_end_time
    ON public.auctions(status, end_time);

CREATE INDEX idx_auctions_current_highest_bidder
    ON public.auctions(current_highest_bidder_id);

CREATE INDEX idx_bids_auction_time
    ON public.bids(auction_id, bid_time DESC);

CREATE INDEX idx_bids_bidder_id
    ON public.bids(bidder_id);

CREATE INDEX idx_orders_buyer_status
    ON public.orders(buyer_id, status);

CREATE INDEX idx_orders_seller_status
    ON public.orders(seller_id, status);

CREATE INDEX idx_transactions_order_id
    ON public.transactions(order_id);

CREATE INDEX idx_transactions_buyer_id
    ON public.transactions(buyer_id);

CREATE INDEX idx_notifications_user_unread_created
    ON public.notifications(user_id, is_read, created_at DESC);

CREATE INDEX idx_email_deliveries_notification_status
    ON public.email_deliveries(notification_id, status);

CREATE INDEX idx_reports_status_created
    ON public.reports(status, created_at DESC);

CREATE INDEX idx_reports_reporter_user
    ON public.reports(reporter_user_id);

CREATE INDEX idx_moderation_actions_admin_created
    ON public.moderation_actions(admin_user_id, created_at DESC);

CREATE INDEX idx_subscriptions_user_status
    ON public.subscriptions(user_id, status);
