alter table nrtm_rpslobject
    rename constraint rpslobject__nrtm_source__object_type__primary_key__from_version__uid to rpslobject__source__type__primary_key__from_version__uid;
alter table nrtm_rpslobject
    rename constraint rpslobject__nrtm_source__object_type__primary_key__to_version__uid to rpslobject__source__type__primary_key__to_version__uid;

---- create above / drop below ----

alter table nrtm_rpslobject
    rename constraint rpslobject__source__type__primary_key__to_version__uid to rpslobject__nrtm_source__object_type__primary_key__to_version__uid;
alter table nrtm_rpslobject
    rename constraint rpslobject__source__type__primary_key__from_version__uid to  rpslobject__nrtm_source__object_type__primary_key__from_version__uid;