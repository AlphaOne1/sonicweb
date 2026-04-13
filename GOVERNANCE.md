<!--  SPDX-FileCopyrightText: © 2004-2025 Greg Wilson
      SPDX-FileCopyrightText: 2026 The SonicWeb contributors.
      SPDX-License-Identifier: CC-BY-4.0
-->

*SonicWeb* Project Governance
=============================

The *SonicWeb* project uses a governance model commonly described as Benevolent
Dictator For Life (BDFL). This document outlines our implementation of this
model.


Terms
-----

_user_: anyone who interacts with the project.

_contributor_: anyone who has contributed to the project in any way and to any
degree. Contributions may include code, documentation, reviews, manual testing,
etc. All contributors are listed in the project's metadata
(e.g., [AUTHORS.md](AUTHORS.md)).

_core contributor_: anyone who has contributed repeatedly and significantly to
the project. Core contributors are recognized by GitHub “Member” badges.

_Benevolent Dictator For Life (BDFL)_: the person who makes decisions when
consensus cannot be reached. The project’s current BDFL is Alexander Adam
(@AlphaOne1).


Decision-Making Process
-----------------------

Contributors try to reach consensus via discussion. When consensus cannot be
reached in a reasonable time, the BDFL takes the following steps:

 1. Creates an issue in the project’s GitHub repository labeled governance with
    a suitable title.

 2. Waits at least 48 hours for contributors to add comments to the issue.

 3. Adds a comment with the final decision and closes the issue.


Changing Governance
-------------------

The BDFL may choose a replacement at any point and cede authority to them. If
this happens, the decision must be announced and recorded in the manner
described above for decisions in the absence of consensus.

The new BDFL must clearly state that they accept their new responsibilities.

The time at which the new BDFL assumes their responsibilities must be clearly
stated.

The BDFL may alternatively decide to move to a Steering Committee governance
model, in which case this document must be replaced with a new description of
roles and responsibilities.


Access to Sensitive Resources
-----------------------------

The following project members have access to sensitive resources (GitHub
Secrets, Repository Settings, Release Keys):

* Alexander Adam (@AlphaOne1) - BDFL / Primary Maintainer

Currently, no other contributors have administrative access to the
build infrastructure or cryptographic keys.


Project Continuity
------------------

To ensure the project's continuity, the following measures are taken:

*GitHub Successor*: A designated successor has been appointed in the GitHub
account settings to take over repository management if the primary maintainer is
incapacitated.

<!-- Currently this is not implemented, uncomment if it is indeed enacted:
*Access Recovery*: Recovery keys for the repository and build infrastructure are
stored in a secure digital vault with emergency access enabled for a trusted
party.
-->

*Forking*: As a fallback, the MPL-2.0 license ensures that the community can
fork and continue the project at any time.
