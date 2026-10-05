BEGIN;

ALTER TABLE public.reports
    ADD CONSTRAINT reports_status_check
        CHECK (status IN ('PENDING','RESOLVED','REJECTED')),
    ADD CONSTRAINT reports_exactly_one_target_check
        CHECK (num_nonnulls(reported_user_id, listing_id, auction_id) = 1);

ALTER TABLE public.moderation_actions
    ADD CONSTRAINT moderation_actions_type_check
        CHECK (action_type IN ('WARN_USER','SUSPEND_LISTING','SUSPEND_SELLER'));


COMMIT;


